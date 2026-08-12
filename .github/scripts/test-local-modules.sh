#!/usr/bin/env bash

set -euo pipefail

readonly repository_dir="$(git rev-parse --show-toplevel)"
readonly temporary_dir="$(mktemp -d)"

cleanup() {
	rm -rf "${temporary_dir}"
}
trap cleanup EXIT

declare -a local_modules=()
while IFS= read -r mod_file; do
	module_dir="$(dirname "${mod_file}")"
	module_path="$(cd "${module_dir}" && GOWORK=off go list -m -f '{{.Path}}')"
	local_modules+=("${module_path}|${module_dir}")
done < <(find "${repository_dir}" -name go.mod -not -path '*/.git/*' -print | sort)

while IFS= read -r mod_file; do
	module_dir="$(dirname "${mod_file}")"
	module_path="$(cd "${module_dir}" && GOWORK=off go list -m -f '{{.Path}}')"
	module_key="${module_path//\//_}"
	temporary_mod="${temporary_dir}/${module_key}.mod"
	temporary_sum="${temporary_dir}/${module_key}.sum"

	cp "${mod_file}" "${temporary_mod}"
	if [[ -f "${module_dir}/go.sum" ]]; then
		cp "${module_dir}/go.sum" "${temporary_sum}"
	fi

	for local_module in "${local_modules[@]}"; do
		dependency_path="${local_module%%|*}"
		dependency_dir="${local_module#*|}"
		if [[ "${dependency_path}" == "${module_path}" ]]; then
			continue
		fi
		if go mod edit -modfile="${temporary_mod}" -json |
			jq -e --arg path "${dependency_path}" 'any(.Require[]?; .Path == $path)' >/dev/null; then
			go mod edit -modfile="${temporary_mod}" \
				-replace="${dependency_path}=${dependency_dir}"
		fi
	done

	echo "Testing ${module_path} against local repository modules"
	(
		cd "${module_dir}"
		GOWORK=off go mod tidy -modfile="${temporary_mod}"
		GOWORK=off go test -v -race -mod=readonly -modfile="${temporary_mod}" ./...
	)
done < <(find "${repository_dir}/cmd" "${repository_dir}/contrib" -name go.mod -print | sort)
