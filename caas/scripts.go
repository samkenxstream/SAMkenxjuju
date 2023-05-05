// Copyright 2019 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package caas

var (
	// JujudStartUpSh is the start script for K8s controller and operator style agents.
	JujudStartUpSh = `
export JUJU_DATA_DIR=%[1]s
export JUJU_TOOLS_DIR=$JUJU_DATA_DIR/%[2]s

mkdir -p $JUJU_TOOLS_DIR
cp /opt/jujud $JUJU_TOOLS_DIR/jujud

%[3]s
`[1:]

	// JujudStartUpAltSh is the start script for K8s operator style agents.
	JujudStartUpAltSh = `
export JUJU_DATA_DIR=%[1]s
export JUJU_TOOLS_DIR=$JUJU_DATA_DIR/%[2]s

mkdir -p $JUJU_TOOLS_DIR
cp %[3]s/jujud $JUJU_TOOLS_DIR/jujud

%[4]s
`[1:]

	// MongoStartupShTemplate is used to generate the start script for mongodb.
	MongoStartupShTemplate = `
args="%s"
ipv6Disabled=$(sysctl net.ipv6.conf.all.disable_ipv6 -n)
if [ $ipv6Disabled -eq 0 ]; then
  args="${args} --ipv6"
fi
exec mongod ${args}
`[1:]

	// JujudCopySh is the start script for K8s operator style agents.
	JujudCopySh = `
cp /opt/jujud %[1]s/jujud

%[2]s
`[1:]

	// APIServerStartUpSh is the start script for the "api-server" container
	// in the controller pod (Pebble running jujud).
	APIServerStartUpSh = `
export JUJU_DATA_DIR=%[1]s
export JUJU_TOOLS_DIR=$JUJU_DATA_DIR/tools

mkdir -p $JUJU_TOOLS_DIR
cp /opt/jujud $JUJU_TOOLS_DIR/jujud

%[2]s

mkdir -p /var/lib/pebble/default/layers
cat > /var/lib/pebble/default/layers/001-jujud.yaml <<EOF
%[3]s
EOF

/opt/pebble run --http :%[4]s --verbose
`[1:]
)
