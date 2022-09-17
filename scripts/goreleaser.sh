#!/usr/bin/env sh
# Wrap goreleaser by using envsubst on .goreleaser.yml.
set -e
this_script="$(basename $0)"
hash envsubst
hash goreleaser
export skip_feature_docker_tags=false
export skip_release=true
if [ "${CIRCLE_BRANCH}" == "" ] ; then
  echo "${this_script} requires the CIRCLE_BRANCH environment variable to be set to the current git branch"
  exit 1
  fi
  if [ "${CIRCLE_BRANCH}" == "testmaster" ] ; then
  echo "${this_script} setting skip_release to false, and skip_feature_docker_tags to true,  because this is the main branch"
export skip_feature_docker_tags=true
export skip_release=false
else
  # Use an adjusted git branch name as an additional docker tag, for feature branches.
  export feature_docker_tag=$(echo "${CIRCLE_BRANCH:0:26}" | sed 's/[^a-zA-Z0-9]/-/g' | sed 's/-\+$//')
  echo "${this_script} also using docker tag ${feature_docker_tag} since ${CIRCLE_BRANCH} is a feature branch"
fi
cat .goreleaser.yml.envsubst |envsubst >.goreleaser.yml
goreleaser $@
if [ $? -eq 0 ] ; then
  echo "${this_script} removing the temporary .goreleaser.yml since goreleaser was successful"
  rm .goreleaser.yml # Keep git clean for additional goreleaser runs
fi
