#!/bin/bash
if [[ -z "$INPUT_VERSION" ]]; then
  echo "Missing polaris version information"
  exit 1
fi

polaris version | grep "$INPUT_VERSION" &> /dev/null
if [ $? == 0 ]; then
   echo "Polaris $INPUT_VERSION is already installed! Exiting gracefully."
   exit 0
fi

INPUT_FILE='polaris_'"$INPUT_VERSION"'_linux_386.tar.gz'

API_URL="https://api.github.com/repos/FairwindsOps/polaris"
RELEASE_DATA=$(curl $API_URL/releases/tags/${INPUT_VERSION})
ASSET_ID=$(echo $RELEASE_DATA | jq -r ".assets | map(select(.name == \"${INPUT_FILE}\"))[0].id")
TAG_VERSION=$(echo $RELEASE_DATA | jq -r ".tag_name" | sed -e "s/^v//" | sed -e "s/^v.//")

if [[ -z "$ASSET_ID" ]]; then
  echo "Could not specified versions asset id"
  exit 1
fi

curl \
  -J \
  -L \
  -H "Accept: application/octet-stream" \
  "$API_URL/releases/assets/$ASSET_ID" \
  -o ${INPUT_FILE}

mkdir polaris

tar -xzf $INPUT_FILE -C polaris

rm $INPUT_FILE

echo $GITHUB_PATH

echo $RUNNER_TOOL_CACHE
ls -la $RUNNER_TOOL_CACHE

echo "polaris" >> $GITHUB_PATH

echo "::set-output name=version::$TAG_VERSION"