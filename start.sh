#!/bin/sh

echo "prepare environment"

echo "start mock.bill.com server"

/sbin/su-exec app /srv/mockserver $@