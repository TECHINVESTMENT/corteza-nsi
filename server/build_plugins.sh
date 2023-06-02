#!/usr/bin/env bash

CURR_DIR=$(pwd)
PLUGIN_DIR=./pkg/plugin/plugin
PLUGIN_DIR_DEST=./plugin

echo ${CURR_DIR}

rm ${CURR_DIR}/${PLUGIN_DIR_DEST}/*

for dir in "${PLUGIN_DIR}"/*; do
    if [ -d "$dir" ]; then
        package=$(basename ${dir})
        echo "> building plugin ${package}"
        cd $dir && go build -o ${CURR_DIR}/${PLUGIN_DIR_DEST}/${package} .
        cd ${CURR_DIR}
    fi
done


