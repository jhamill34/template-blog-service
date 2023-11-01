# Building your images

In local dev you need a dot env file to hold information for interacting with ECR and 
knowing which namespace to set (i.e. your domain name)

```bash 
cp ./build/templates/env ./build/.env
vim ./build/.env
```

If ECR has not been setup for this domain yet run the following to create your repositories.

```bash 
./build/provision.sh
```

The following will build your code into docker images, tag them, and then push them up to
ECR

```bash 
./build/build_all.sh
./build/distribute.sh
```


