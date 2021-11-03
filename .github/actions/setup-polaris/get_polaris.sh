#!/bin/bash
if [[ -z "$INPUT_VERSION" ]]; then
  echo "Missing polaris version information"
  exit 1
fi
POLARIS_URL=https://github.com/FairwindsOps/polaris/releases/download/$INPUT_VERSION/polaris_linux_amd64.tar.gz
polaris version | grep "$INPUT_VERSION" &> /dev/null
if [ $? == 0 ]; then
   echo "Polaris $INPUT_VERSION is already installed! Exiting gracefully."
   exit 0
else
  echo "Installing polaris to path from " $POLARIS_URL
fi
TARGET_FILE="polaris.tar.gz"
curl -LJ -o $TARGET_FILE $POLARIS_URL
mkdir polaris
tar -xzf $TARGET_FILE -C polaris
rm $TARGET_FILE
echo "polaris" >> $GITHUB_PATH
echo "::set-output name=version::$INPUT_VERSION"