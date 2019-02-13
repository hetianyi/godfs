#!/bin/sh

echo -e "+-------------------------------------+"
echo -e "|              >>godfs<<              |"
echo -e "| focus project on github: \033[0m           |"
echo -e "| \033[4mhttps://github.com/hetianyi/godfs\033[0m   |"
echo -e "+-------------------------------------+"

if [ ! -e "./bin" ];then
    echo "please make first!"
    exit 1
fi

if [ "$1"x = ""x ];then
    echo "no installation directory specified."
    exit 2
fi

echo "install to directory: $1"
if [ ! -e "$1" ];then
    echo "installation directory not found, auto create it."
    mkdir -p "$1"
fi

\cp -rf bin $1
\cp -rf conf $1
rm -f /usr/bin/client
cd /usr/bin/ && ln -s $1/bin/client client
echo "install success!"
