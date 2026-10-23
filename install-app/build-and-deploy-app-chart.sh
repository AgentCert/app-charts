#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_CONTEXT="$(cd "$SCRIPT_DIR/.." && pwd)"

IMAGE_NAMESPACE="${IMAGE_NAMESPACE:-agentcert}"
IMAGE_NAME="${IMAGE_NAME:-agentcert-install-app}"
IMAGE_TAG="${IMAGE_TAG:-ci-$(date +%Y%m%d%H%M%S)}"
MINIKUBE_PROFILE="${MINIKUBE_PROFILE:-minikube}"
TAG_LATEST="${TAG_LATEST:-true}"
NO_CACHE="${NO_CACHE:-true}"
AGENTCERT_ENV_FILE="${AGENTCERT_ENV_FILE:-$BUILD_CONTEXT/../.env}"
PUSH=false
DEPLOY=false
LOCAL_REQUESTED=0
LOCAL_MODE=false

SOCKSHOP_VALUES_FILE="$BUILD_CONTEXT/charts/sock-shop/values.yaml"

IMAGE_REPO="${IMAGE_NAMESPACE}/${IMAGE_NAME}"
PRIMARY_IMAGE="${IMAGE_REPO}:${IMAGE_TAG}"
LATEST_IMAGE="${IMAGE_REPO}:latest"

usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Build install-app locally by default. Publishing and Minikube loading are opt-in.

Options:
  --push         Publish to the configured registry
  --deploy       Load into Minikube and update the selected .env
  --env-file PATH  Configuration file (default: monorepo .env)
  --source-dir PATH  install-app source directory
  --allow-build-cache  Reuse unchanged Docker layers
  --local        Build only (default)
  --local-mode   Build with local-friendly sock-shop tracing settings
                 (temporary override: tracing.disableSleuth=true, tracing.zipkinHost="")
  --help         Show this help
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --local) LOCAL_REQUESTED=1; shift ;;
        --push) PUSH=true; shift ;;
        --deploy) DEPLOY=true; shift ;;
        --allow-build-cache) NO_CACHE=false; shift ;;
        --env-file|--source-dir)
            [[ $# -ge 2 && -n "$2" && "$2" != --* ]] || { echo "ERROR: $1 requires a path" >&2; exit 2; }
            if [[ "$1" == --env-file ]]; then AGENTCERT_ENV_FILE="$2"; else SCRIPT_DIR="$(cd "$2" && pwd)"; fi
            shift 2
            ;;
        --local-mode)
            LOCAL_MODE=true
            shift
            ;;
        --help|-h)
            usage
            exit 0
            ;;
        *)
            printf '[ERROR] Unknown option: %s\n' "$1" >&2
            usage
            exit 1
            ;;
    esac
done
if [[ "${LOCAL_REQUESTED}" -eq 1 ]] && [[ "${DEPLOY}" == true || "${PUSH}" == true ]]; then
    echo 'ERROR: --local cannot be combined with --push or --deploy.' >&2
    exit 2
fi


BUILD_CONTEXT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SOCKSHOP_VALUES_FILE="${BUILD_CONTEXT}/charts/sock-shop/values.yaml"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=../../scripts/lib/registry.sh
source "${REPO_ROOT}/scripts/lib/registry.sh"
registry_load "${AGENTCERT_ENV_FILE}"
# This legacy wrapper publishes its explicit timestamp/latest tags.
ACE_IMAGE_TAG=""

info() {
    printf '\n[INFO] %s\n' "$1"
}

success() {
    printf '[OK] %s\n' "$1"
}

warn() {
    printf '[WARN] %s\n' "$1"
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || {
        printf '[ERROR] Required command not found: %s\n' "$1" >&2
        exit 1
    }
}

prepare_local_mode_overrides() {
    if [[ "$LOCAL_MODE" != "true" ]]; then
        return 0
    fi

    if [[ ! -f "$SOCKSHOP_VALUES_FILE" ]]; then
        warn "sock-shop values not found at $SOCKSHOP_VALUES_FILE; skipping local-mode chart overrides"
        return 0
    fi

    info "Applying local-mode tracing overrides to sock-shop chart (temporary)"
    LOCAL_BUILD_CONTEXT="$(mktemp -d)"
    trap 'rm -rf "$LOCAL_BUILD_CONTEXT"' EXIT
    cp -a "$BUILD_CONTEXT/charts" "$BUILD_CONTEXT/install-app" "$LOCAL_BUILD_CONTEXT/"
    BUILD_CONTEXT="$LOCAL_BUILD_CONTEXT"
    SCRIPT_DIR="$BUILD_CONTEXT/install-app"
    SOCKSHOP_VALUES_FILE="$BUILD_CONTEXT/charts/sock-shop/values.yaml"

    sed -i 's|^\([[:space:]]*zipkinHost:[[:space:]]*\).*|\1""|' "$SOCKSHOP_VALUES_FILE"
    sed -i 's|^\([[:space:]]*disableSleuth:[[:space:]]*\).*|\1true|' "$SOCKSHOP_VALUES_FILE"

    success "Local-mode overrides applied: tracing.disableSleuth=true, tracing.zipkinHost=\"\""
}

upsert_env_value() {
    local key="$1"
    local value="$2"
    local escaped
    escaped=$(printf '%s' "$value" | sed 's/[&/\\]/\\&/g')

    if grep -q -E "^${key}=" "$AGENTCERT_ENV_FILE"; then
        sed -i "s/^${key}=.*/${key}=${escaped}/" "$AGENTCERT_ENV_FILE"
    else
        printf '\n%s=%s\n' "$key" "$value" >> "$AGENTCERT_ENV_FILE"
    fi
}

build_image() {
    local build_args=(docker_build_resolved "$SCRIPT_DIR/Dockerfile" -t "$PRIMARY_IMAGE")
    if [[ "$NO_CACHE" == "true" ]]; then
        build_args+=(--no-cache)
    fi
    build_args+=("$BUILD_CONTEXT")

    info "Building ${PRIMARY_IMAGE}"
    "${build_args[@]}"

    if [[ "$TAG_LATEST" == "true" ]]; then
        info "Tagging ${PRIMARY_IMAGE} as ${LATEST_IMAGE}"
        docker tag "$PRIMARY_IMAGE" "$LATEST_IMAGE"
    fi

    success "Docker build completed"
}

load_into_minikube() {
    info "Loading ${PRIMARY_IMAGE} into ${MINIKUBE_PROFILE}"
    minikube -p "$MINIKUBE_PROFILE" image load "$PRIMARY_IMAGE"

    if [[ "$TAG_LATEST" == "true" ]]; then
        info "Loading ${LATEST_IMAGE} into ${MINIKUBE_PROFILE}"
        minikube -p "$MINIKUBE_PROFILE" image load "$LATEST_IMAGE"
    fi

    success "Images loaded into minikube"
}



update_agentcert_env() {
    if [[ ! -f "$AGENTCERT_ENV_FILE" ]]; then
        warn "AgentCert .env not found at ${AGENTCERT_ENV_FILE}; skipping env update"
        return 0
    fi

    info "Updating AgentCert .env with ${PRIMARY_IMAGE}"
    upsert_env_value "INSTALL_APPLICATION_IMAGE" "$PRIMARY_IMAGE"
    success "AgentCert .env updated: INSTALL_APPLICATION_IMAGE=${PRIMARY_IMAGE}"
}

show_result() {
    printf '\nBuilt image: %s\n' "$PRIMARY_IMAGE"
    if [[ "$TAG_LATEST" == "true" ]]; then
        printf 'Alias image: %s\n' "$LATEST_IMAGE"
    fi
    if [[ "$DEPLOY" == true ]]; then
        printf 'Minikube profile: %s\n' "$MINIKUBE_PROFILE"
        printf 'Updated .env: %s\n' "$AGENTCERT_ENV_FILE"
    fi
}

main() {
    require_cmd docker
    if [[ "$DEPLOY" == true ]]; then require_cmd minikube; fi

    if [[ "$PUSH" == true ]]; then
        registry_has_credentials || { echo 'ERROR: --push requires registry credentials' >&2; exit 1; }
        registry_login
    fi
    prepare_local_mode_overrides

    build_image
    if [[ "$PUSH" == true ]]; then
        local dst
        dst="$(registry_ref "$PRIMARY_IMAGE")"
        [[ "$dst" == "$PRIMARY_IMAGE" ]] || docker tag "$PRIMARY_IMAGE" "$dst"
        docker push "$dst"
        if [[ "$TAG_LATEST" == true ]]; then
            dst="$(registry_ref "$LATEST_IMAGE")"
            [[ "$dst" == "$LATEST_IMAGE" ]] || docker tag "$LATEST_IMAGE" "$dst"
            docker push "$dst"
        fi
    fi
    if [[ "$DEPLOY" == true ]]; then
        load_into_minikube
        update_agentcert_env
    fi
    show_result
}

main "$@"