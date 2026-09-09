#!/bin/bash
set -e


do_install(){
	if [ -f "/etc/systemd/system/atlas-runner.service" ]; then
		systemctl stop atlas-runner
		systemctl disable atlas-runner
		systemctl daemon-reload
	fi
}

do_upgrade(){
	if [ -f "/etc/systemd/system/atlas-runner.service" ]; then
		systemctl stop atlas-runner
		systemctl disable atlas-runner
		systemctl daemon-reload
	fi
}

do_other(){

}

case "$1" in
  install)
    do_install
    ;;
  upgrade)
    do_upgrade
    ;;
  *)
    do_other
    ;;
esac
