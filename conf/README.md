# Bindata

To update the bindata within this project you need to install the go-bindata
utitlity, follow the steps below to get this utility and to update the included
bindata. We are not covering within this steps how to setup an Go environment.


## Prepare

To get the required utility execute the following commands, generally this have
to be done only once.

```
go get -u github.com/jteeuwen/go-bindata
go install github.com/jteeuwen/go-bindata/...
```


## Update

To update the current bindata within the resulting binary you need to execute
the following command always if one of these configuration files changes.
Execute this command always in the root directory of this project.

```
go-bindata -o=modules/bindata/bindata.go -ignore="\\.DS_Store|README.md" -pkg=bindata conf/...
```
