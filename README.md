# Online-Boutique

Online Boutique is a cloud-first microservices demo application. The application is a web-based e-commerce app where users can browse items, add them to the cart, and purchase them.

Adapted from https://github.com/GoogleCloudPlatform/microservices-demo/tree/main (all of the sevices that weren't written in Go have been ported to Go and aRPC.)

## Architecture

**Online Boutique** is composed of 11 microservices written in different languages that talk to each other over aRPC.

[![Architecture of
microservices](./architecture-diagram.png)](./architecture-diagram.png)


## Build Docker Images and Push them to DockerHub

```bash
# You may need to change the $USER in `build_images.sh`,
# and set $UPDATE_ARPC to use latest/pinned aRPC version. 
bash build_images.sh # run `docker login -u $username` first to connect to DockerHub.

# To test the build locally
go build -o /tmp/test_build ./cmd/main.go
```

## Run Bookinfo Applicaton

```bash
kubectl apply -Rf ./kubernetes/apply
kubectl get pods

# Test (Home Handler)
curl http://10.96.88.88/ -d "user_id=test"

# Checkout Handler
curl -X POST http://10.96.88.88/cart/checkout -d "email=test@example.com" -d "street_address=123 Main St" -d "zip_code=98101" -d "city=Seattle" -d "state=WA" -d "country=USA" -d "credit_card_number=4111111111111111" -d "credit_card_expiration_month=12" -d "credit_card_expiration_year=2025" -d "credit_card_cvv=123" -d "user_id=test"

# wrk
./utils/wrk -c 1 -t 1 http://10.96.88.88/ -d 30s -L

# Destroy
kubectl delete pv,pvc,sa,all --all
```

## Open Jaeger UI

```bash
kubectl port-forward svc/jaeger 16686:16686

xdg-open http://localhost:16686
```
