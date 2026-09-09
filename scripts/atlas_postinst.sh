#!/bin/bash
set -e


do_configure(){
	if [ -f "/etc/systemd/system/atlas.service" ]; then
		systemctl start atlas
		systemctl enable atlas
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
