#!/bin/bash
set -eux
input="/etc/netdevice.conf"

if [ ! -f $input ]; then
	echo "File /etc/netdevice.conf doesn't exist."
	exit
fi

source common.sh

dev_number=$(jq -r '.deviceList | length' ${input})
if (( dev_number<1 ));then
	echo "No network device configured in ${input}"
	exit
fi

for (( i=0; i<dev_number; i++ ))
do
	pci_addr=$(jq -r ".deviceList[${i}].pci_addr" ${input})
	int_name=$(jq -r ".deviceList[${i}].int_name" ${input})
	driver_model=$(jq -r ".deviceList[${i}].driver_model" ${input})	

	add_udev_rule $pci_addr $int_name

	if [ ! -z "${pci_addr}" ];then
		if [ ! -z "$driver_model" ] && [ ! -z "$int_name" ]; then
			# set device driver model
			devlink dev eswitch set pci/${pci_addr} mode ${driver_model}
			ip link set ${int_name} up
		fi
	fi

	for feature_name in "${ethtool_features[@]}"
	do
		feature_value=$(jq -r ".deviceList[${i}].ethtool.${feature_name}" ${input})
		if [ ! -z "${feature_value}" ];then
			ethtool -K ${int_name} ${feature_name} ${feature_value}
		fi
	done
done
