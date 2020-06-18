# Kong setup
This repo shows a setup of using [Kong as ingress controller in Kubernetes](https://github.com/Kong/kubernetes-ingress-controller).
It is configured with one endpoint that has rate limiting applied based on the client id for an [OAuth Bearer token
in the request `Authorization` header](https://tools.ietf.org/html/rfc6750#section-2.1).

The client id is read from the token using a [token introspection request](https://tools.ietf.org/html/rfc7662#section-2.1)
to a specified endpoint.

The following steps shows how to run it locally, using [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/):

* Build Docker image for Kong, including custom plugins, and make sure the [image is available in Minikube](https://github.com/kubernetes/minikube/blob/0c616a6b42b28a1aab8397f5a9061f8ebbd9f3d9/README.md#reusing-the-docker-daemon):
    ```console
    $ minikube start
    $ eval $(minikube docker-env)
    $ docker build -t kong-with-local-plugin .
    ```
* Run Kong in k8s:
    ```console
    $ helm repo add kong https://charts.konghq.com
    $ helm repo update
    $ helm init
    $ helm install -f kong-override.yaml --name kong kong/kong
    $ export PROXY_IP=$(minikube service kong-kong-proxy --url | head -1)
    ```
* Setup echo-server (from [here](https://github.com/Kong/kubernetes-ingress-controller/blob/master/docs/guides/getting-started.md)):
    ```console
    $ kubectl apply -f https://bit.ly/echo-service
    ```
* First, configure `introspection_endpoint` and `introspection_client_credentials` in `ingress.yaml`, then add ingress rule with Kong plugins configured:
    ```console
    $ kubectl apply -f ingress.yaml
    ```
  
To test it, make some requests and check the returned rate limiting headers: 
```console
$ curl -i ${PROXY_IP}/foo # without authorization, rate limiting defaults to client IP
$ curl -i ${PROXY_IP}/foo -H 'Authorization: Bearer <token>' # with authorization, rate limiting will use client id from valid token
```
