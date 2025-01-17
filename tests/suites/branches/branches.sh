run_branch() {
	# Echo out to ensure nice output to the test suite.
	echo

	# The following ensures that a bootstrap juju exists
	file="${TEST_DIR}/test-branches.log"
	ensure "branches" "${file}"

	juju branch | check 'Active branch is "master"'

	juju deploy mongodb --channel 3.6/stable

	wait_for "mongodb" "$(idle_condition "mongodb")"

	juju add-branch test-branch
	juju branch | check 'Active branch is "test-branch"'

	juju config mongodb replicaset=newset --branch test-branch
	check_config_command "juju config mongodb replicaset --branch test-branch" "newset"
	check_config_command "juju config mongodb replicaset --branch master" "myset"

	juju commit test-branch | check 'Active branch set to "master"'
	check_config_command "juju config mongodb replicaset" "newset"

	# Clean up!
	destroy_model "branches"
}

test_branch() {
	if [ -n "$(skip 'test_branch')" ]; then
		echo "==> SKIP: Asked to skip branch tests"
		return
	fi

	(
		set_verbosity

		cd .. || exit

		run "run_branch"
	)
}

# The check function reads stdout until a new-line is detected.
# This does not work for us, because config does not include a new line at the end
# of the output. Hence this function.
check_config_command() {
	local want got

	want=${2}

	got=$(eval "${1}")

	OUT=$(echo "${got}" | grep -E "${want}" || true)
	if [ -z "${OUT}" ]; then
		echo "" >&2
		# shellcheck disable=SC2059
		printf "$(red \"Expected\"): ${want}\n" >&2
		# shellcheck disable=SC2059
		printf "$(red \"Recieved\"): ${got}\n" >&2
		echo "" >&2
		exit 1
	fi
}
