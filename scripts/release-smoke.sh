#!/usr/bin/env bash
set -euo pipefail

archive="${1:-}"
if [[ -z "${archive}" ]]; then
  echo "usage: $0 <release-archive.tar.gz|release-archive.zip>" >&2
  exit 2
fi
if [[ ! -f "${archive}" ]]; then
  echo "release archive not found: ${archive}" >&2
  exit 1
fi

workdir="${SOHA_RELEASE_SMOKE_WORKDIR:-$(mktemp -d)}"
cleanup() {
  if [[ -z "${SOHA_RELEASE_SMOKE_WORKDIR:-}" ]]; then
    rm -rf "${workdir}"
  fi
}
trap cleanup EXIT

extract_dir="${workdir}/extract"
mkdir -p "${extract_dir}"
case "${archive}" in
  *.tar.gz)
    tar -xzf "${archive}" -C "${extract_dir}"
    ;;
  *.zip)
    unzip -q "${archive}" -d "${extract_dir}"
    ;;
  *)
    echo "unsupported release archive format: ${archive}" >&2
    exit 1
    ;;
esac

soha_bin="$(find "${extract_dir}" -type f \( -name soha -o -name soha.exe \) -print -quit)"
if [[ -z "${soha_bin}" ]]; then
  echo "release archive does not contain a soha binary" >&2
  exit 1
fi
chmod +x "${soha_bin}"

version_json="${workdir}/version.json"
"${soha_bin}" version --json > "${version_json}"
grep -q '"version"' "${version_json}"
grep -q '"commit"' "${version_json}"

mcp_json="${workdir}/mcp.json"
"${soha_bin}" mcp install \
  --profile release-smoke \
  --command "${soha_bin}" \
  --ai-client-id release-smoke \
  --ai-client "Release Smoke" \
  --skill-id k8s-sre > "${mcp_json}"
grep -q "\"command\": \"${soha_bin}\"" "${mcp_json}"
grep -q '"release-smoke"' "${mcp_json}"
grep -q '"k8s-sre"' "${mcp_json}"

home="${workdir}/home"
runtime_skills="${workdir}/runtime-skills"
mkdir -p "${home}" "${runtime_skills}"

skills_source="${workdir}/skills-source"
skills_arg="k8s-sre"
if [[ -n "${SOHA_SKILLS_ARTIFACT:-}" ]]; then
  if [[ ! -f "${SOHA_SKILLS_ARTIFACT}" ]]; then
    echo "skills artifact not found: ${SOHA_SKILLS_ARTIFACT}" >&2
    exit 1
  fi
  artifact_dir="$(cd "$(dirname "${SOHA_SKILLS_ARTIFACT}")" && pwd)"
  artifact_name="$(basename "${SOHA_SKILLS_ARTIFACT}")"
  artifact_path="${artifact_dir}/${artifact_name}"
  checksum_file="${SOHA_SKILLS_ARTIFACT_SHA256:-${SOHA_SKILLS_ARTIFACT}.sha256}"
  if [[ ! -f "${checksum_file}" ]]; then
    echo "skills artifact checksum not found: ${checksum_file}" >&2
    exit 1
  fi
  checksum_dir="$(cd "$(dirname "${checksum_file}")" && pwd)"
  checksum_path="${checksum_dir}/$(basename "${checksum_file}")"
  (
    cd "${artifact_dir}"
    shasum -a 256 -c "${checksum_path}"
  )
  skills_extract="${workdir}/skills-artifact"
  mkdir -p "${skills_extract}"
  tar -xzf "${artifact_path}" -C "${skills_extract}"
  skills_source="${skills_extract}/soha-skills"
  skills_arg="all"
else
  mkdir -p "${skills_source}/k8s-sre"
  cat > "${skills_source}/k8s-sre/SKILL.md" <<'EOF'
---
name: k8s-sre
description: Release smoke fixture.
---

# K8s SRE
EOF
fi

add_output="${workdir}/add-all.txt"
HOME="${home}" \
SOHA_CONFIG="${workdir}/soha-config.json" \
SOHA_SKILLS_DIR="${runtime_skills}" \
"${soha_bin}" add all \
  --dry-run \
  --profile release-smoke \
  --server https://mcp.opensoha.com \
  --source "${skills_source}" \
  --command "${soha_bin}" \
  --skills "${skills_arg}" \
  --runtime-skill-dest "${runtime_skills}" > "${add_output}"

for target in Codex Claude Cursor Kiro Gemini Antigravity Trae; do
  grep -q "Would update ${target} MCP config" "${add_output}"
  grep -q "Would install ${target} Soha skill package" "${add_output}"
done
grep -q "Would install Soha runtime skills" "${add_output}"

echo "release smoke passed for ${archive}"
