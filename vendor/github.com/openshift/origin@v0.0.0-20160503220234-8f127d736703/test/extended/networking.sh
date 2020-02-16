#!/bin/bash

# This script runs the networking e2e tests. See CONTRIBUTING.adoc for
# documentation.

set -o errexit
set -o nounset
set -o pipefail

if [[ -n "${OPENSHIFT_VERBOSE_OUTPUT:-}" ]]; then
  set -o xtrace
  export PS4='+ \D{%b %d %H:%M:%S} $(basename ${BASH_SOURCE}):${LINENO} ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
fi

# Ensure that subshells inherit bash settings (specifically xtrace)
export SHELLOPTS

OS_ROOT=$(dirname "${BASH_SOURCE}")/../..
source "${OS_ROOT}/hack/util.sh"
source "${OS_ROOT}/hack/common.sh"
source "${OS_ROOT}/hack/lib/log.sh"
source "${OS_ROOT}/hack/lib/util/environment.sh"
os::log::install_errexit

NETWORKING_DEBUG=${NETWORKING_DEBUG:-false}

# These strings filter the available tests.
NETWORKING_E2E_FOCUS="${NETWORKING_E2E_FOCUS:-etworking|Services}"
NETWORKING_E2E_SKIP="${NETWORKING_E2E_SKIP:-}"

DEFAULT_SKIP_LIST=(
  # Skip tests that require secrets.  Secrets are not supported by
  # dind without docker >= 1.10.
  "Networking should function for intra-pod"
  "openshift router"

  # DNS inside container fails in CI but works locally
  "should provide Internet connection for containers"
)

CLUSTER_CMD="${OS_ROOT}/hack/dind-cluster.sh"

# Control variable to limit unnecessary cleanup
DIND_CLEANUP_REQUIRED=0

function copy-container-files() {
  local source_path=$1
  local base_dest_dir=$2

  for container_name in "${CONTAINER_NAMES[@]}"; do
    local dest_dir="${base_dest_dir}/${container_name}"
    if [[ ! -d "${dest_dir}" ]]; then
      mkdir -p "${dest_dir}"
    fi
    sudo docker cp "${container_name}:${source_path}" "${dest_dir}"
  done
}

function save-container-logs() {
  local base_dest_dir=$1

  os::log::info "Saving container logs"

  for container_name in "${CONTAINER_NAMES[@]}"; do
    local dest_dir="${base_dest_dir}/${container_name}"
    if [[ ! -d "${dest_dir}" ]]; then
      mkdir -p "${dest_dir}"
    fi
    container_log_file=/tmp/systemd.log.gz
    sudo docker exec -t "${container_name}" bash -c "journalctl -xe | \
gzip > ${container_log_file}"
    sudo docker cp "${container_name}:${container_log_file}" "${dest_dir}"
  done
}

function save-artifacts() {
  local name=$1
  local config_root=$2

  os::log::info "Saving cluster configuration"

  local dest_dir="${ARTIFACT_DIR}/${name}"

  local config_source="${config_root}/openshift.local.config"
  local config_dest="${dest_dir}/openshift.local.config"
  mkdir -p "${config_dest}"
  cp -r ${config_source}/* ${config_dest}/

  copy-container-files "/etc/hosts" "${dest_dir}"
}

function deploy-cluster() {
  local name=$1
  local plugin=$2
  local isolation=$3
  local log_dir=$4

  os::log::info "Launching a docker-in-docker cluster for the ${name} plugin"
  export OPENSHIFT_NETWORK_PLUGIN="${plugin}"
  export OPENSHIFT_CONFIG_ROOT="${BASETMPDIR}/${name}"
  export OPENSHIFT_NETWORK_ISOLATION="${isolation}"
  # Images have already been built
  export OPENSHIFT_DIND_BUILD_IMAGES=0
  DIND_CLEANUP_REQUIRED=1

  local exit_status=0

  # Restart instead of start to ensure that an existing test cluster is
  # always torn down.
  if ${CLUSTER_CMD} restart; then
    if ! ${CLUSTER_CMD} wait-for-cluster; then
      exit_status=1
    fi
  else
    exit_status=1
  fi

  save-artifacts "${name}" "${OPENSHIFT_CONFIG_ROOT}"

  return "${exit_status}"
}

# Any non-zero exit code from any test run invoked by this script
# should increment TEST_FAILURE so the total count of failed test runs
# can be returned as the exit code.
TEST_FAILURES=0
function test-osdn-plugin() {
  local name=$1
  local plugin=$2
  local isolation=$3

  os::log::info "Targeting ${name} plugin: ${plugin}"

  local log_dir="${LOG_DIR}/${name}"
  mkdir -p "${log_dir}"

  local failed=

  if deploy-cluster "${name}" "${plugin}" "${isolation}" "${log_dir}"; then
    os::log::info "Running networking e2e tests against the ${name} plugin"
    export TEST_REPORT_FILE_NAME="${name}-junit"

    if ! TEST_REPORT_FILE_NAME=networking_${name}_${isolation} \
         run-extended-tests "${OPENSHIFT_CONFIG_ROOT}" "${log_dir}/test.log"; then
      failed=1
      os::log::error "e2e tests failed for plugin: ${plugin}"
    fi
  else
    failed=1
    os::log::error "Failed to deploy cluster for plugin: {$name}"
  fi

  if [[ -n "${failed}" ]]; then
    TEST_FAILURES=$((TEST_FAILURES + 1))
  fi

  save-container-logs "${log_dir}"

  os::log::info "Shutting down docker-in-docker cluster for the ${name} plugin"
  ${CLUSTER_CMD} stop
  DIND_CLEANUP_REQUIRED=0
  rmdir "${OPENSHIFT_CONFIG_ROOT}"
}


function join { local IFS="$1"; shift; echo "$*"; }

function run-extended-tests() {
  local config_root=$1
  local log_path=${2:-}

  local focus_regex="${NETWORKING_E2E_FOCUS}"
  local skip_regex="${NETWORKING_E2E_SKIP}"

  if [[ -z "${skip_regex}" ]]; then
      skip_regex=$(join '|' "${DEFAULT_SKIP_LIST[@]}")
  fi

  export KUBECONFIG="${config_root}/openshift.local.config/master/admin.kubeconfig"
  export EXTENDED_TEST_PATH="${OS_ROOT}/test/extended"

  local test_args="--test.v '--ginkgo.skip=${skip_regex}' \
'--ginkgo.focus=${focus_regex}' ${TEST_EXTRA_ARGS}"

  if [[ "${NETWORKING_DEBUG}" = 'true' ]]; then
    local test_cmd="dlv exec ${TEST_BINARY} -- ${test_args}"
  else
    local test_cmd="${TEST_BINARY} ${test_args}"
  fi

  if [[ -n "${log_path}" ]]; then
    test_cmd="${test_cmd} | tee ${log_path}"
  fi

  pushd "${EXTENDED_TEST_PATH}/networking" > /dev/null
    eval "${test_cmd}; "'exit_status=${PIPESTATUS[0]}'
  popd > /dev/null

  return ${exit_status}
}

CONFIG_ROOT="${OPENSHIFT_CONFIG_ROOT:-}"
case "${CONFIG_ROOT}" in
  dev)
    CONFIG_ROOT="${OS_ROOT}"
    ;;
  dind)
    CONFIG_ROOT="/tmp/openshift-dind-cluster/\
${OPENSHIFT_INSTANCE_PREFIX:-openshift}"
    if [[ ! -d "${CONFIG_ROOT}" ]]; then
      os::log::error "OPENSHIFT_CONFIG_ROOT=dind but dind cluster not found"
      os::log::info  "To launch a cluster: hack/dind-cluster.sh start"
      exit 1
    fi
    ;;
  *)
    if [[ -n "${CONFIG_ROOT}" ]]; then
      CONFIG_FILE="${CONFIG_ROOT}/openshift.local.config/master/admin.kubeconfig"
      if [[ ! -f "${CONFIG_FILE}" ]]; then
        os::log::error "${CONFIG_FILE} not found"
        exit 1
      fi
    fi
    ;;
esac

TEST_EXTRA_ARGS="$@"

if [[ "${OPENSHIFT_SKIP_BUILD:-false}" = "true" ]] &&
     [[ -n $(os::build::find-binary extended.test) ]]; then
  os::log::warn "Skipping rebuild of test binary due to OPENSHIFT_SKIP_BUILD=true"
else
  hack/build-go.sh test/extended/extended.test
fi
TEST_BINARY="${OS_ROOT}/$(os::build::find-binary extended.test)"

os::log::info "Starting 'networking' extended tests"
if [[ -n "${CONFIG_ROOT}" ]]; then
  os::log::info "CONFIG_ROOT=${CONFIG_ROOT}"
  # Run tests against an existing cluster
  run-extended-tests "${CONFIG_ROOT}"
else
  # For each plugin, run tests against a test-managed cluster

  # Use a unique instance prefix to ensure the names of the test dind
  # containers will not clash with the names of non-test containers.
  export OPENSHIFT_INSTANCE_PREFIX="nettest"
  # TODO(marun) Discover these names instead of hard-coding
  CONTAINER_NAMES=(
    "${OPENSHIFT_INSTANCE_PREFIX}-master"
    "${OPENSHIFT_INSTANCE_PREFIX}-node-1"
    "${OPENSHIFT_INSTANCE_PREFIX}-node-2"
  )

  os::util::environment::setup_tmpdir_vars "test-extended/networking"
  reset_tmp_dir

  os::log::start_system_logger

  os::log::info "Building docker-in-docker images"
  ${CLUSTER_CMD} build-images

  # Ensure cleanup on error
  ENABLE_SELINUX=0
  function cleanup-dind {
    local exit_code=$?
    if [[ "${DIND_CLEANUP_REQUIRED}" = "1" ]]; then
      os::log::info "Shutting down docker-in-docker cluster"
      ${CLUSTER_CMD} stop || true
    fi
    enable-selinux || true
    if [[ "${TEST_FAILURES}" = "0" ]]; then
      os::log::info "No test failures were detected"
    else
      os::log::error "${TEST_FAILURES} plugin(s) failed one or more tests"
    fi
    # Return non-zero for either command or test failures
    if [[ "${exit_code}" = "0" ]]; then
      exit_code="${TEST_FAILURES}"
    else
      os::log::error "Exiting with code ${exit_code}"
    fi
    exit $exit_code
  }
  trap "exit" INT TERM
  trap "cleanup-dind" EXIT

  # Docker-in-docker is not compatible with selinux
  disable-selinux

  # Ignore deployment errors for a given plugin to allow other plugins
  # to be tested.
  test-osdn-plugin "subnet" "redhat/openshift-ovs-subnet" "false" || true

  # Avoid unnecessary go builds for subsequent deployments
  export OPENSHIFT_SKIP_BUILD=true

  test-osdn-plugin "multitenant" "redhat/openshift-ovs-multitenant" "true" || true
fi
