# client-go-crd-bug

> follow the below steps to reproduce the issue documented [here](https://github.com/kubernetes/client-go/issues/276).

## Start Minikube

```sh
$ minikube start
```

## Build Program

```sh
$ mkdir -p $GOPATH/src/github.com/frankgreco/
$ cd $GOPATH/src/github.com/frankgreco/
$ git clone git@github.com:frankgreco/client-go-crd-bug.git
$ cd client-go-crd-bug
$ glide install
$ go build
```

## Create CRD Manually

```sh
$ kubectl apply -f crd.yml
$ kubect apply -f example.yml
```

## Start Program

```sh
$ ./client-go-crd-bug --kubeconfig=~/.kube/config
adding crd named example
```

## Reproduce Bug

```sh
$ sed -i -e 's/ApiFoo/APIFoo/g' main.go
$ go build
$ ./client-go-crd-bug --kubeconfig=~/.kube/config
E0927 20:37:00.072532   26614 reflector.go:201] github.com/frankgreco/client-go-crd-bug/main.go:96: Failed to list *main.APIFoo: no kind "ApiFooList" is registered for version "bar.io/v1"
```