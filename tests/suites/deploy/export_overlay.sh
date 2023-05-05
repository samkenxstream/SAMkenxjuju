run_cmr_bundles_export_overlay() {
	echo

	file="${TEST_DIR}/test-cmr-bundles-export-overlay.log"

	ensure "cmr-bundles-test-export-overlay" "${file}"

	juju add-user bar
	juju deploy ./tests/suites/deploy/bundles/bundle-with-overlays/easyrsa.yaml

	OUT=$(juju export-bundle 2>&1)
	echo "${OUT}"

	# ensure that overlay.yaml is exported
	echo "${OUT}" | grep -- "--- # overlay.yaml"

	juju add-model test1

	echo -n 'my-include' >example.log
	cat >overlay.yaml <<EOT
applications:
  etcd:
    annotations:
      raw: include-file://example.log
      enc: include-base64://example.log
EOT

	juju deploy ./tests/suites/deploy/bundles/bundle-with-overlays/easyrsa-etcd.yaml --overlay overlay.yaml
	OUT=$(juju export-bundle 2>&1)
	echo "${OUT}"

	# did the annotations and overlay get exported?
	echo "${OUT}" | grep -- "--- # overlay.yaml"
	echo "${OUT}" | check "enc: bXktaW5jbHVkZQ=="
	echo "${OUT}" | check "raw: my-include"

	destroy_model "cmr-bundles-test-export-overlay"
	destroy_model "test1"
}

test_cmr_bundles_export_overlay() {
	if [ "$(skip 'test_cmr_bundles_export_overlay')" ]; then
		echo "==> TEST SKIPPED: CMR bundle deploy tests"
		return
	fi

	(
		set_verbosity

		cd .. || exit

		run "run_cmr_bundles_export_overlay"
	)
}
