#!/bin/bash
set -e


do_configure(){
	if [ -f "/etc/systemd/system/atlas-runner.service" ]; then
		systemctl start atlas-runner
		systemctl enable atlas-runner
		systemctl daemon-reload
	fi
}

do_other(){

}

case "$1" in
  configure)
    do_configure
    ;;
  *)
    do_other
    ;;
esac
