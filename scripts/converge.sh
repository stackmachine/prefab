#!/bin/bash
set -x

pushd /vagrant/software

./../stack postgresql/manifest.json redis/manifest.json \
	python/manifest.json

# DEVELOPMENT
popd

pushd /vagrant
sh scripts/installgo.sh
popd