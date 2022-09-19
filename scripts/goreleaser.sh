#!/usr/bin/env sh
# Wrap goreleaser by using envsubst on .goreleaser.yml,
# and creating a temporary git tag.

function cleanup {
if [ "${CIRCLE_TAG}" == "" ] ; then
  echo "${this_script} deleting git tag ${temporary_git_tag} for goreleaser"
  unset GORELEASER_CURRENT_TAG
  git tag -d ${temporary_git_tag}
fi
}

set -eE # errexit and errtrace
trap 'cleanup' ERR
this_script="$(basename $0)"
if [ "${CIRCLE_BRANCH}" == "" ] ; then
  echo "${this_script} requires the CIRCLE_BRANCH environment variable, which is not set"
  exit 1
  fi
  hash envsubst
hash goreleaser
if [ "${TMPDIR}" == "" ] ; then
  export TMPDIR="/tmp"
  echo "${this_script} temporarily set the TMPDIR environment variable to ${TMPDIR}, used for a temporary GOBIN environment variable"
fi
if [ "${CIRCLE_TAG}" == "" ] ; then
  last_git_tag="$(git describe --tags --abbrev=0 2>/dev/null)"
  if [ "${last_git_tag}" == "" ] ; then
    echo "${this_script} is unable to determine the last git tag using: git describe --tags --abbrev=0"
    exit 1
  fi
  if [ "$(git config user.email)" == "" ] ; then
    # git will use this env var as its user.email.
    # git tag -m is used in case tags are manually pushed by accident,
    # however git tag -m requires an email.
    export EMAIL='goreleaser_ci@fairwinds.com'
    echo "${this_script} using ${EMAIL} temporarily as the git user.email"
  fi
  temporary_git_tag=$(echo "${last_git_tag}" | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')-rc
  echo "${this_script} creating git tag ${temporary_git_tag} for goreleaser, the last real tag is ${last_git_tag}"
  # The -f is included to overwrite existing tags, perhaps from previous CI jobs.
  git tag -f -m "temporary local tag for goreleaser" ${temporary_git_tag}
  export GORELEASER_CURRENT_TAG=${temporary_git_tag}
  else
  export GORELEASER_CURRENT_TAG=${CIRCLE_TAG}
fi
echo "${this_script} using git tag ${GORELEASER_CURRENT_TAG}"
export skip_feature_docker_tags=false
export skip_release=true
# CIRCLE_BRANCH is used because its safer than relying on CIRCLE_TAG to only be set during main/master merges.
if [ "${CIRCLE_BRANCH}" == "master" ] ; then
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
cleanup

