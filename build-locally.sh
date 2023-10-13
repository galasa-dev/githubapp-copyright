#!/usr/bin/env bash

#
# Copyright contributors to the Galasa project
#
# SPDX-License-Identifier: EPL-2.0
#

# Where is this script executing from ?
BASEDIR=$(dirname "$0");pushd $BASEDIR 2>&1 >> /dev/null ;BASEDIR=$(pwd);popd 2>&1 >> /dev/null
# echo "Running from directory ${BASEDIR}"
export ORIGINAL_DIR=$(pwd)
cd "${BASEDIR}"


#--------------------------------------------------------------------------
#
# Set Colors
#
#--------------------------------------------------------------------------
bold=$(tput bold)
underline=$(tput sgr 0 1)
reset=$(tput sgr0)

red=$(tput setaf 1)
green=$(tput setaf 76)
white=$(tput setaf 7)
tan=$(tput setaf 202)
blue=$(tput setaf 25)

#--------------------------------------------------------------------------
#
# Headers and Logging
#
#--------------------------------------------------------------------------
underline() { printf "${underline}${bold}%s${reset}\n" "$@"
}
h1() { printf "\n${underline}${bold}${blue}%s${reset}\n" "$@"
}
h2() { printf "\n${underline}${bold}${white}%s${reset}\n" "$@"
}
debug() { printf "${white}%s${reset}\n" "$@"
}
info() { printf "${white}➜ %s${reset}\n" "$@"
}
success() { printf "${green}✔ %s${reset}\n" "$@"
}
error() { printf "${red}✖ %s${reset}\n" "$@"
}
warn() { printf "${tan}➜ %s${reset}\n" "$@"
}
bold() { printf "${bold}%s${reset}\n" "$@"
}
note() { printf "\n${underline}${bold}${blue}Note:${reset} ${blue}%s${reset}\n" "$@"
}

#-----------------------------------------------------------------------------------------                   
# Functions
#-----------------------------------------------------------------------------------------                   
function usage {
    info "Syntax: build-locally.sh"
}

#--------------------------------------------------------------------------
# 
# Main script logic
#
#--------------------------------------------------------------------------

#-----------------------------------------------------------------------------------------                   
# Process parameters
#-----------------------------------------------------------------------------------------                   
build_type=""

while [ "$1" != "" ]; do
    case $1 in
        -h | --help )           usage
                                exit
                                ;;
        * )                     error "Unexpected argument $1"
                                usage
                                exit 1
    esac
    shift
done



function clean {
    h2 "Cleaning the binaries out..."
    if [[ "${build_type}" != "clean" ]]; then
        success "No need to clean up."
    else
        make clean
        rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to build binary executable copyright checker programs. rc=${rc}" ; exit 1 ; fi
        success "Binaries cleaned up - OK"
    fi
}

#--------------------------------------------------------------------------
#
# Build the executables
#
#--------------------------------------------------------------------------
function build_executables {

    h2 "Building new binaries..."
    set -o pipefail # Fail everything if anything in the pipeline fails. Else we are just checking the 'tee' return code.
    mkdir -p ${BASEDIR}/build
    export GITHUB_AUTH_KEY=${BASEDIR}/build/key.pem
    make all | tee ${BASEDIR}/build/compile-log.txt
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to build binary executable copyright checker programs. rc=${rc}. See log at ${BASEDIR}/build/compile-log.txt" ; exit 1 ; fi
    success "New binaries built - OK"
}

function build_container_image {
    h2 "Building container image..."

    cmd="docker build -t githubapp-copyright:latest -f Dockerfile --build-arg=dockerRepository=docker.io/library/ ."
    info "Command is $cmd"
    $cmd
    rc=$? ; if [[ "${rc}" != "0" ]]; then error "Failed to build container image. rc=${rc}. See log at ${BASEDIR}/build/compile-log.txt" ; exit 1 ; fi

    success "Built container image OK."
}

#--------------------------------------------------------------------------
h1 "Building the copyright checker tool"

clean
build_executables
build_container_image

success "OK"
