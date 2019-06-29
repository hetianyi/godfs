#!/bin/sh

# this is a simple script for development.
rm -f /usr/bin/godfs
ln -s $PWD/bin/godfs /usr/bin/godfs
echo "install finish!"
