#!/bin/bash

set -x

case "$1" in
    -s | --setup)
        vni="$2"
        echo "Setup debug env with vni [$vni]"
        # Add interface and netns
        ip l add s$vni type bridge
        ip l add kr_$vni type veth peer name "kr_$vni_p"
        ip l add t$vni type vxlan id $vni dstport 4789
        ip netns add nr_$vni

        # Set interface and up
        ip l set kr_$vni master s$vni
        ip l set kr_$vni_p netns nr_$vni
        ip l set t$vni master s$vni
        ip l set s$vni up
        ip l set kr_$vni up
        ip l set t$vni up
        ip netns exec nr_$vni ip l set lo up
        ip netns exec nr_$vni ip l set kr_$vni_p up
        ;;
    -t | --teardown)
        vni="$2"
        echo "Teardown debug env with vni [$vni]"
        ip netns del nr_$vni
        ip l del s$vni
        ;;
    -h | --help)
        echo "Usage:"
        echo "$0 -s [vni] Setup debug env with specific vni"
        echo "$0 -t [vni] Teardown debug env with specific vni"
        ;;
esac
