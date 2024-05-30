#!/bin/bash

BUILD_PATH=$1

CLI_PATH=$BUILD_PATH/cli

echo "Building baseos image ..."
$CLI_PATH image build --source-path $BUILD_PATH --dockerfile $BUILD_PATH/Dockerfile.baseos --image-tag baseos:bookworm --log-level debug

echo "Pushing baseos image ..."
$CLI_PATH image push --image-tag baseos:bookworm --log-level debug

echo "Building node image ..."
$CLI_PATH image build --source-path $BUILD_PATH --dockerfile $BUILD_PATH/Dockerfile.node --image-tag node:20.x --log-level debug

echo "Pushing node image ..."
$CLI_PATH image push --image-tag node:20.x --log-level debug

echo "Building initial node image ..."
$CLI_PATH image build --source-path $BUILD_PATH --dockerfile $BUILD_PATH/Dockerfile.initial-node --image-tag initial-node:20.x-x86_64 --log-level debug

echo "Pushing initial node image ..."
$CLI_PATH image push --image-tag initial-node:20.x-x86_64 --log-level debug

echo "Done."
