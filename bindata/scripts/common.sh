#!/bin/bash

UDEV_RULE_FILE="/etc/udev/rules.d/10-persistent-net.rules"

append_to_file(){
	content="$1"
	file_name="$2"

	if ! test -f "$file_name";then
		echo "$content" > "$file_name"
	else
		if ! grep -Fxq "$content" "$file_name";then
      			echo "$content" >> "$file_name"
		fi
	fi
}

add_udev_rule(){
	pci_addr=$1
	name=$2

	udev_data_line="SUBSYSTEM==\"net\", ACTION==\"add\", DRIVERS==\"?*\", KERNELS==\"$pf_pci\", NAME=\"$1\""
	append_to_file "$udev_data_line" "$UDEV_RULE_FILE"
}

ethtool_features=(
hw-tc-offload
hw-tc-offload
)
