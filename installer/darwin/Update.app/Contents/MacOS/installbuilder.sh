#!/bin/sh

os_version=`uname -r`
machine_platform=`uname -p`
if [ "${os_version:0:1}" == "6" ];then
    executable="none"
elif [ "${machine_platform}" == "i386" ];then
    executable="osx-intel"
else
    executable="none"
fi

if [ "$executable" == "none" ]; then
    echo "The current OS X version is not supported"
    exit 1
fi
            
        
if [ "${1}" == --help ] || [ "`id -u 2>/dev/null`" == "0" ];then
    "`dirname \"${0}\"`/$executable" "$@"
else
    "`dirname \"${0}\"`/The Whitecat Create Agent" $executable "$@"
fi
                