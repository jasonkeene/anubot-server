#!/bin/bash

set -e

passed[0]="ᕕ( ᐛ )ᕗ"
passed[1]="(╯✧∇✧)╯"
passed[2]="(づ｡◕‿‿◕｡)づ"
passed[3]="(๑˃̵ᴗ˂̵)و"

failed[0]="щ(ಥДಥщ)"
failed[1]="ヽ(｀⌒´メ)ノ"
failed[2]="(屮ಠ益ಠ)屮"
failed[3]="ヽ(#ﾟДﾟ)ﾉ┌┛"
failed[4]="(ノಠ益ಠ)ノ彡┻━┻"

function passed {
    local rand=$[ $RANDOM % 4 ]
    echo
    echo -e "  \033[32;48;5;2m  ${passed[$rand]}                    \033[0m"
    echo -e "  \033[30;48;5;2m  ${passed[$rand]} ALL TESTS PASSED!  \033[0m"
    echo -e "  \033[32;48;5;2m  ${passed[$rand]}                    \033[0m"
    echo
}

function failed {
    local rand=$[ $RANDOM % 5 ]
    echo
    echo -e "  \033[31;48;5;1m  ${failed[$rand]}               \033[0m"
    echo -e "  \033[30;48;5;1m  ${failed[$rand]} TEST FAILED!  \033[0m"
    echo -e "  \033[31;48;5;1m  ${failed[$rand]}               \033[0m"
    echo
}

function check_status {
    if [ $1 -ne 0 ]; then
        failed
    else
        passed
    fi
}
function handle_exit {
    status=$?
    for exit_func in "${exit_funcs[@]}"; do
        $exit_func
    done
    check_status $status
}
trap handle_exit EXIT

function checkpoint {
    echo
    echo -e "  \033[38;5;104m$@\033[0m"
    echo
}

function header {
    echo
    echo -e "  \033[30;48;5;104m                             \033[0m"
    echo -e "  \033[30;48;5;104m   ___ ___ _ _| |_ ___| |_   \033[0m"
    echo -e "  \033[30;48;5;104m  | .'|   | | | . | . |  _|  \033[0m"
    echo -e "  \033[30;48;5;104m  |__,|_|_|___|___|___|_|    \033[0m"
    echo -e "  \033[30;48;5;104m              test suite     \033[0m"
    echo -e "  \033[30;48;5;104m                             \033[0m"
}

function tools {
    if [ "$1" = "ci" ]; then
        checkpoint installing tools
        scripts/install_tools.sh
    fi
}

function start_postgres {
    if [ "$1" = "ci" ]; then
        checkpoint starting postgres
        /etc/init.d/postgresql start
        function wait_for_postgres {
            checkpoint waiting for postgres to listen
            until pg_isready -U postgres -h localhost; do
                sleep 0.1;
            done
        }
    else
        checkpoint starting postgres in a container
        PG_CONTAINER=$(docker run --rm -d -p 5432:5432 -e POSTGRES_PASSWORD=pass postgres:latest)
        echo $PG_CONTAINER
        export ANUBOT_TEST_POSTGRES="postgres://postgres:pass@localhost:5432/postgres?sslmode=disable"
        function teardown {
            checkpoint tearing down postgres container
            docker kill $PG_CONTAINER
        }
        exit_funcs[0]=teardown
        function wait_for_postgres {
            checkpoint waiting for postgres to listen
            until docker run --rm --link $PG_CONTAINER:pg postgres:latest pg_isready -U postgres -h pg; do
                sleep 0.1;
            done
        }
    fi
}

function build {
    checkpoint building binaries
    scripts/build.sh
}

function migrate_postgres {
    checkpoint migrating postgres database
    migrate -path=store/migrations -url="$ANUBOT_TEST_POSTGRES" up
}

function test {
    checkpoint running tests
    non_vendor_pkgs=$(go list ./... | grep -v /vendor/)

    go test -race $non_vendor_pkgs
}

function clean {
    checkpoint running cleaners
    non_vendor_dirs=$(ls -d */ | grep -v vendor/)

    echo goimports
    goimports -w $non_vendor_dirs

    echo gofmt
    gofmt -s -w $non_vendor_dirs

    echo misspell
    misspell -w $non_vendor_dirs
}

function lint {
    checkpoint running linters
    non_vendor_pkgs=$(go list ./... | grep -v /vendor/)
    non_vendor_dirs=$(ls -d */ | grep -v vendor/)

    echo go vet
    go vet $non_vendor_pkgs

    echo deadcode
    deadcode $non_vendor_dirs

    echo golint
    for pkg in $non_vendor_pkgs; do
        dir=${pkg#github.com/jasonkeene/anubot-server/}
        go_files=$(ls "$dir" | grep -E "\.go$")

        if [ -z "$go_files" ]; then
            continue
        fi

        main_pkg_files=$(echo "$go_files" | grep -vE "_test.go$" | cat)
        if [ -n "$main_pkg_files" ]; then
            prefixed_main_pkg_files=$(echo "$main_pkg_files" | sed -e "s|^|$dir/|")
            if [ -n "$prefixed_main_pkg_files" ]; then
                golint_files=$(echo "$prefixed_main_pkg_files" |
                                xargs grep -L '// Code generated by "stringer' |
                                cat)
                golint $golint_files
            fi
        fi

        test_pkg_files=$(echo "$go_files" | grep -E "_test.go$" | cat)
        if [ -n "$test_pkg_files" ]; then
            prefixed_test_pkg_files=$(echo "$test_pkg_files" | sed -e "s|^|$dir/|")
            if [ -n "$prefixed_test_pkg_files" ]; then
                golint_files=$(echo $prefixed_test_pkg_files |
                                xargs grep -L '// Code generated by "stringer' |
                                cat)
                golint $golint_files
            fi
        fi
    done

    echo aligncheck
    aligncheck $non_vendor_pkgs

    echo structcheck
    structcheck $non_vendor_pkgs

    echo varcheck
    varcheck $non_vendor_pkgs

    echo ineffassign
    for dir in $non_vendor_dirs; do
        ineffassign $dir
    done

    echo interfacer
    interfacer $non_vendor_pkgs

    echo unconvert
    unconvert -v -apply $non_vendor_pkgs

    echo gosimple
    gosimple $non_vendor_pkgs

    echo staticcheck
    staticcheck $non_vendor_pkgs

    echo unused
    unused $non_vendor_pkgs

    echo errcheck
    cat <<EOF > errcheckexclude
(*database/sql.Tx).Rollback
(*database/sql.Stmt).Close
(*database/sql.Rows).Close
EOF
    errcheck -exclude errcheckexclude $non_vendor_pkgs
    rm errcheckexclude
}

function main {
    header
    start_postgres $1
    build
    wait_for_postgres
    migrate_postgres
    test
    tools $1
    clean
    lint
}
main $@
