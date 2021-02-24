#!/bin/bash
if [[ -z "$INPUT_VERSION" ]]; then
  echo "Missing polaris version information"
  exit 1
fi
polaris version | grep "$INPUT_VERSION" &> /dev/null
if [ $? == 0 ]; then
   echo "Polaris $INPUT_VERSION is already installed! Exiting gracefully."
   exit 0
else
  echo "Installing polaris to path."
fi
TARGET_FILE="polaris.tar.gz"
curl -LJ -o $TARGET_FILE 'https://github.com/FairwindsOps/polaris/releases/download/'"$INPUT_VERSION"'/polaris_'"$INPUT_VERSION"'_linux_386.tar.gz'
mkdir polaris
tar -xzf $TARGET_FILE -C polaris
rm $TARGET_FILE
echo "polaris" >> $GITHUB_PATH
echo "::set-output name=version::$INPUT_VERSION"