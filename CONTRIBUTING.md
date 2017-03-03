## Steps for Contributing:

1. Fork the repo master branch (or feature branch if adding to another feature).
1. Add or modify unit tests to properly test your change.  Make sure code coverage does not decrease.
1. If your change effects the usage, update README to reflect the change.
1. Make sure you have proper GoDoc documentation for all Public interfaces, structs, and functions.
1. Submit Pull Request of your fork against master branch (or the branch of origin).

## Getting the Source
Use git clones to get the source from your fork

```sh
$ git clone https://github.com/mygithubid/ctrace-go.git
```

or from the main repo.

```sh
$ git clone https://github.com/Nordstrom/ctrace-go.git
```

## Initialize Project
Use make dependencies to initialize the project

```sh
make dependencies
```

## Building and Testing Project
To lint the project run make lint

```sh
make lint
```

To test the project run make test

```sh
make test
```

To run bench marks run make bench

```sh
make bench
```
