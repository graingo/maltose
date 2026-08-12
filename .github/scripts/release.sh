#!/usr/bin/env bash

set -euo pipefail

readonly ROOT_MODULE="github.com/graingo/maltose"
readonly RELEASE_BRANCH="${RELEASE_BRANCH:-master}"
readonly OBSERVABILITY_MODULE="${OBSERVABILITY_MODULE:-./contrib/observability}"
readonly VERSION="${VERSION:?VERSION is required}"
readonly COMMAND="${1:?release command is required}"

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

leaf_modules=()
while IFS= read -r module_dir; do
	[[ -n "$module_dir" ]] && leaf_modules+=("$module_dir")
done < <(printf '%s\n' "${LEAF_MODULES:-}" | sed '/^$/d')
if ((${#leaf_modules[@]} == 0)); then
	echo "::error::LEAF_MODULES must contain at least one module."
	exit 1
fi

tag_exists() {
	git show-ref --verify --quiet "refs/tags/$1"
}

module_tag() {
	local module_dir="$1"
	printf '%s/%s' "${module_dir#./}" "$VERSION"
}

commit_staged_changes() {
	local message="$1"
	if git diff --cached --quiet; then
		echo "-> No release changes need to be committed."
		return
	fi
	git commit -m "$message"
}

wait_for_module() {
	local module_path="$1"
	local module_version="$2"
	local attempt

	echo "-> Waiting for ${module_path}@${module_version} to become downloadable."
	for attempt in {1..30}; do
		if GOWORK=off go mod download "${module_path}@${module_version}"; then
			return
		fi
		sleep 2
	done

	echo "::error::${module_path}@${module_version} was not downloadable after 60 seconds."
	exit 1
}

verify_tag_on_release_branch() {
	local tag_name="$1"
	local tag_sha
	tag_sha="$(git rev-list -n 1 "$tag_name")"
	if ! git merge-base --is-ancestor "$tag_sha" "origin/${RELEASE_BRANCH}"; then
		echo "::error::Tag ${tag_name} points outside origin/${RELEASE_BRANCH}."
		exit 1
	fi
}

verify_module_dependency() {
	local tag_name="$1"
	local module_dir="$2"
	local dependency="$3"
	local dependency_version="$4"
	local go_mod_path="${module_dir#./}/go.mod"

	if ! git show "${tag_name}:${go_mod_path}" |
		awk -v dependency="$dependency" -v version="$dependency_version" '
      $1 == dependency && $2 == version { found = 1 }
      END { exit found ? 0 : 1 }
    '; then
		echo "::error::Tag ${tag_name} does not require ${dependency}@${dependency_version}."
		exit 1
	fi
}

validate_context() {
	if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
		echo "::error::Version must use the vX.Y.Z format."
		exit 1
	fi

	if [[ "${GITHUB_REF_NAME:-}" != "$RELEASE_BRANCH" ]]; then
		echo "::error::Releases must run from ${RELEASE_BRANCH}; current ref is ${GITHUB_REF_NAME:-unknown}."
		exit 1
	fi

	git fetch origin "$RELEASE_BRANCH" --tags

	local local_sha remote_sha
	local_sha="$(git rev-parse HEAD)"
	remote_sha="$(git rev-parse "origin/${RELEASE_BRANCH}")"
	if [[ "$local_sha" != "$remote_sha" ]]; then
		echo "::error::The workflow is not running from the latest ${RELEASE_BRANCH} commit."
		echo "::error::local=${local_sha} remote=${remote_sha}"
		exit 1
	fi

	local root_exists=false
	local leaves_exist=false
	local observability_exists=false
	local leaf_count=0
	local root_release_sha=""
	local leaf_release_sha=""
	local module_dir tag_name

	if tag_exists "$VERSION"; then
		root_exists=true
	fi

	for module_dir in "${leaf_modules[@]}"; do
		tag_name="$(module_tag "$module_dir")"
		if tag_exists "$tag_name"; then
			((leaf_count += 1))
		fi
	done

	if ((leaf_count != 0 && leaf_count != ${#leaf_modules[@]})); then
		echo "::error::Only part of the first-level module tags exist; refusing an ambiguous recovery."
		exit 1
	fi
	if ((leaf_count == ${#leaf_modules[@]})); then
		leaves_exist=true
	fi

	local observability_tag
	observability_tag="$(module_tag "$OBSERVABILITY_MODULE")"
	if tag_exists "$observability_tag"; then
		observability_exists=true
	fi

	if [[ "$root_exists" != true && ("$leaves_exist" == true || "$observability_exists" == true) ]]; then
		echo "::error::A child module tag exists without the root ${VERSION} tag."
		exit 1
	fi
	if [[ "$leaves_exist" != true && "$observability_exists" == true ]]; then
		echo "::error::The observability tag exists before all of its dependencies were published."
		exit 1
	fi

	if [[ "$root_exists" == true ]]; then
		verify_tag_on_release_branch "$VERSION"
		root_release_sha="$(git rev-list -n 1 "$VERSION")"
		if ! git show "${VERSION}:version.go" | grep -q "VERSION = \"${VERSION}\""; then
			echo "::error::Tag ${VERSION} does not contain the matching version.go value."
			exit 1
		fi
	fi

	if [[ "$leaves_exist" == true ]]; then
		for module_dir in "${leaf_modules[@]}"; do
			tag_name="$(module_tag "$module_dir")"
			verify_tag_on_release_branch "$tag_name"
			verify_module_dependency "$tag_name" "$module_dir" "$ROOT_MODULE" "$VERSION"
			if [[ -z "$leaf_release_sha" ]]; then
				leaf_release_sha="$(git rev-list -n 1 "$tag_name")"
			elif [[ "$(git rev-list -n 1 "$tag_name")" != "$leaf_release_sha" ]]; then
				echo "::error::First-level module tags do not point to the same release commit."
				exit 1
			fi
		done
		if ! git merge-base --is-ancestor "$root_release_sha" "$leaf_release_sha"; then
			echo "::error::First-level module tags precede the root ${VERSION} tag."
			exit 1
		fi
	fi

	if [[ "$observability_exists" == true ]]; then
		local observability_release_sha
		observability_release_sha="$(git rev-list -n 1 "$observability_tag")"
		verify_tag_on_release_branch "$observability_tag"
		verify_module_dependency "$observability_tag" "$OBSERVABILITY_MODULE" "$ROOT_MODULE" "$VERSION"
		verify_module_dependency "$observability_tag" "$OBSERVABILITY_MODULE" \
			"${ROOT_MODULE}/contrib/metric/otlpmetric" "$VERSION"
		verify_module_dependency "$observability_tag" "$OBSERVABILITY_MODULE" \
			"${ROOT_MODULE}/contrib/trace/otlptrace" "$VERSION"
		if ! git merge-base --is-ancestor "$leaf_release_sha" "$observability_release_sha"; then
			echo "::error::The observability tag precedes its first-level dependency tags."
			exit 1
		fi
	fi

	local mode=publish
	if [[ "$root_exists" == true && "$leaves_exist" == true && "$observability_exists" == true ]]; then
		mode=complete
	fi

	if [[ "$root_exists" == true && "$leaves_exist" != true ]]; then
		if [[ "$local_sha" != "$root_release_sha" ]]; then
			echo "::error::master advanced after the root tag was published; refusing to mix unrelated commits into recovery."
			exit 1
		fi
	elif [[ "$leaves_exist" == true && "$observability_exists" != true ]]; then
		if [[ "$local_sha" != "$leaf_release_sha" ]]; then
			echo "::error::master advanced after first-level tags were published; refusing an ambiguous recovery."
			exit 1
		fi
	fi

	{
		echo "mode=${mode}"
		echo "root_exists=${root_exists}"
		echo "leaves_exist=${leaves_exist}"
		echo "observability_exists=${observability_exists}"
	} >>"${GITHUB_OUTPUT:?GITHUB_OUTPUT is required}"

	echo "-> Release state: root=${root_exists}, leaves=${leaves_exist}, observability=${observability_exists}."
}

validate_workspace() {
	echo "-> Testing the root module against the current checkout."
	GOWORK=off go test -race ./...
	GOWORK=off govulncheck ./...

	# Test one nested module graph at a time. Activating every module in one
	# go.work merges unrelated transitive dependency graphs and can produce
	# conflicts that downstream consumers never encounter.
	.github/scripts/test-local-modules.sh
}

publish_root() {
	sed -i "s/^[[:space:]]*VERSION = \".*\"/\tVERSION = \"${VERSION}\"/" version.go
	if ! grep -q "VERSION = \"${VERSION}\"" version.go; then
		echo "::error::Failed to update version.go to ${VERSION}."
		exit 1
	fi

	git add version.go
	commit_staged_changes "chore(release): prepare ${VERSION} root module"
	git tag "$VERSION"
	git push --atomic origin "HEAD:${RELEASE_BRANCH}" "refs/tags/${VERSION}"
}

publish_leaves() {
	wait_for_module "$ROOT_MODULE" "$VERSION"

	local module_dir module_path
	for module_dir in "${leaf_modules[@]}"; do
		module_path="${module_dir#./}"
		echo "-> Preparing first-level module: ${module_path}"
		(
			cd "$module_dir"
			# Published modules must resolve the tagged root module instead of a
			# repository-local path. Local source composition belongs to go.work.
			go mod edit -dropreplace="$ROOT_MODULE"
			go mod edit -require="${ROOT_MODULE}@${VERSION}"
			GOWORK=off go mod tidy
			GOWORK=off go mod tidy -diff
			GOWORK=off go test -race -mod=readonly ./...
			GOWORK=off govulncheck ./...
		)
		git add "${module_dir}/go.mod" "${module_dir}/go.sum"
	done

	commit_staged_changes "chore(release): prepare ${VERSION} first-level modules"

	local tags=()
	for module_dir in "${leaf_modules[@]}"; do
		module_path="${module_dir#./}"
		git tag "${module_path}/${VERSION}"
		tags+=("refs/tags/${module_path}/${VERSION}")
	done
	git push --atomic origin "HEAD:${RELEASE_BRANCH}" "${tags[@]}"
}

publish_observability() {
	wait_for_module "$ROOT_MODULE" "$VERSION"
	wait_for_module "${ROOT_MODULE}/contrib/metric/otlpmetric" "$VERSION"
	wait_for_module "${ROOT_MODULE}/contrib/trace/otlptrace" "$VERSION"

	(
		cd "$OBSERVABILITY_MODULE"
		go mod edit -dropreplace="$ROOT_MODULE"
		go mod edit -dropreplace="${ROOT_MODULE}/contrib/metric/otlpmetric"
		go mod edit -dropreplace="${ROOT_MODULE}/contrib/trace/otlptrace"
		go mod edit -require="${ROOT_MODULE}@${VERSION}"
		go mod edit -require="${ROOT_MODULE}/contrib/metric/otlpmetric@${VERSION}"
		go mod edit -require="${ROOT_MODULE}/contrib/trace/otlptrace@${VERSION}"
		GOWORK=off go mod tidy
		GOWORK=off go mod tidy -diff
		GOWORK=off go test -race -mod=readonly ./...
		GOWORK=off govulncheck ./...
	)

	git add "${OBSERVABILITY_MODULE}/go.mod" "${OBSERVABILITY_MODULE}/go.sum"
	commit_staged_changes "chore(release): prepare ${VERSION} observability module"

	local tag_name
	tag_name="$(module_tag "$OBSERVABILITY_MODULE")"
	git tag "$tag_name"
	git push --atomic origin "HEAD:${RELEASE_BRANCH}" "refs/tags/${tag_name}"
}

validate_standalone_modules() {
	local module_dir
	for module_dir in "${leaf_modules[@]}" "$OBSERVABILITY_MODULE"; do
		echo "-> Validating published dependency graph: ${module_dir#./}"
		(
			cd "$module_dir"
			GOWORK=off go mod tidy -diff
			GOWORK=off go test -race -mod=readonly ./...
			GOWORK=off govulncheck ./...
		)
	done
}

validate_published_modules() {
	local modules=("$ROOT_MODULE")
	local module_dir module_path module_json module_source

	for module_dir in "${leaf_modules[@]}" "$OBSERVABILITY_MODULE"; do
		module_path="${ROOT_MODULE}/${module_dir#./}"
		modules+=("$module_path")
	done

	for module_path in "${modules[@]}"; do
		wait_for_module "$module_path" "$VERSION"
		module_json="$(GOWORK=off go mod download -json "${module_path}@${VERSION}")"
		module_source="$(jq -er '.Dir' <<<"$module_json")"
		echo "-> Compile-checking published module: ${module_path}@${VERSION}"
		(
			cd "$module_source"
			GOWORK=off go mod tidy -diff
			# Compile every package and test without executing tests that require
			# external services. Full tests ran before the tags were published.
			GOWORK=off go test -race -run '^$' -mod=readonly ./...
		)
	done
}

case "$COMMAND" in
context)
	validate_context
	;;
workspace)
	validate_workspace
	;;
root)
	publish_root
	;;
leaves)
	publish_leaves
	;;
observability)
	publish_observability
	;;
standalone)
	validate_standalone_modules
	;;
published)
	validate_published_modules
	;;
*)
	echo "::error::Unknown release command: ${COMMAND}"
	exit 1
	;;
esac
