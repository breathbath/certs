# Certificate issuing and redirect service (AcmeReverseProxy)

The main goal of this project is to enable https traffic from multiple domains to configured targets through `AcmeReverseProxy`. It allows to issue LetsEncrypt TLS certificates on demand and forward traffic to the configrued targets individually per each host.

## Use case
- you have a running webserver under yourdomain.com
- you want to white-label your service to allow access from customer domains, so a customer visiting client1.domain.com should land on your webserver and get individual look and feel of your website

With the traditional webserver setup you have following issues:
- you need to add customer domains to your webserver to make them available, but often traditional webservers don't allow dynamic host configuration, so you should write application code for managing webserver configuration
- all configuration changes should require webserver restarts, so addition of a new domain would potentially affect the stability of the running webserver
- for all domains you should issue a TLS certificates at runtime
- once certficates expire, you should automatically prolong them while serving incoming requests
- new certificates installations would also require webserver restarts

AcmeReverseProxy solves those problems:
- existing traffic to your webserver is not affected by new customers
- already configured TLS traffic is not affected by new customers
- TLS certificates are issued/prolonged at runtime with the first incoming request so no static webserver configuration is needed
- your webserver consumes TLS traffic naturally. you can add additional logic for client specific requests as AcmeReverseProxy keeps the host and IP of the original request in headers 

### Prerequsites
- Domain pointing to the running AcmeReverseProxy (store it in APP_DOMAIN env variable). If you configure domain for AcmeReverseProxy with CloudFlare, disable proxy mode.
- VPS with free 80 and 443 ports and installed Docker

### How does AcmeReverseProxy work
- Upon start AcmeReverseProxy opens ports 80 and 443 for listening
- If it's started for the first time, it adds APP_DOMAIN to the list of supported domains. So visit APP_DOMAIN for the first time, to issue a TLS certificate for this domain which is needed for the management of the list of supported targets. AcmeReverseProxy will use Acme protocol to issue certificate with LetsEncrypt. Certificate issuing service will call port 80 of AcmeReverseProxy to validate certificate host (APP_DOMAIN in this case)
- Copy the address of your target webserver where all requests will be proxied to e.g. https://yourserver.com
- Now you can call the /add-domain endpoint to register a custom domain of your client, e.g. whitelabel.yourclient.com: 
```
curl https://{APP_DOMAIN}/add-domain?domain=whitelabel.yourclient.com&target=https://yourserver.com -H X-Auth-Key=YOUR_RANODM_KEY -XPOST
```
- Ask your client to add CNAME entry to his domain DNS e.g. `whitelabel.yourclient.com` should have a CNAME pointing to APP_DOMAIN
- Visit `https://whitelabel.yourclient.com`. Your DNS resolver will find the canonical name of this domain (AcmeReverseProxy domain stored in APP_DOMAIN env variable) and will resolve it to an IP of AcmeReverseProxy's VPS
- So your user will land on the AcmeReverseProxy server under https://whitelabel.yourclient.com (on port 443)
- AcmeReverseProxy will look for whitelabel.yourclient.com in the list of configured targets and will find https://yourserver.com. It will now issue a TLS certificate with LetsEncrypt and again use port 80 of AcmeReverseProxy server to validate the host whitelabel.yourclient.com. 
- Once the certificate is issued, AcmeReverseProxy will proxy the request to the configured target, e.g. https://yourserver.com. It will also add `X-Original-Host` header with the initial host (`whitelabel.yourclient.com`) and `X-Forwarded-For` with the original IP of the client (as the original IP will be replaced by the IP of the AcmeReverseProxy server).
- At the end your request to https://whitelabel.yourclient.com will be served by the server at https://yourserver.com.

## Getting started
- Build docker file or use [an existing docker image](https://github.com/breathbath/certs/pkgs/container/certs)
- Autogenerate some password and store id in YOUR_RAND_KEY env variable
- Store the domain of AcmeReverseProxy server in APP_DOMAIN env variable
- Store the domain of your client in CLIENT_DOMAIN env variable
- Store the url of your target server in TARGET_SERVER env variable
- Start docker:
```
sudo docker run -d -p 80:80 -p 443:443 -e AUTH_KEY=$YOUR_RAND_KEY -e APP_DOMAIN=$APP_DOMAIN -v $(pwd)/.certs:/app/.certs ghcr.io/breathbath/certs:master
```
- Visit $APP_DOMAIN to issue a certificate
- Execute
```
curl https://$APP_DOMAIN/add-domain?domain=$CLIENT_DOMAIN&target=$TARGET_SERVER -H X-Auth-Key=$YOUR_RAND_KEY -XPOST
```
- Visit $CLIENT_DOMAIN so you should see the contents of $TARGET_SERVER

## Configuration parameters
- AUTH_KEY to authenticate requests for managing supported domains and reverse proxy targets
- APP_DOMAIN the domain name of `AcmeReverseProxy` service and a corresponding CNAME value of customer domains
- STORAGE_PATH file path to store data

## Endpoints

### Add client domain and associate it with a target URL where requests should go to
```
curl https://$APP_DOMAIN/add-domain?domain=$CLIENT_DOMAIN&target=$TARGET_SERVER -H X-Auth-Key=$YOUR_RAND_KEY -XPOST
```

### Remove client domain and de-associate it with a target URL 
```
curl https://$APP_DOMAIN/remove-domain?domain=$CLIENT_DOMAIN -H X-Auth-Key=$YOUR_RAND_KEY -XDELETE
```