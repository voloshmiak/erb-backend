# erb-backend

How to launch:

Authenticate with GCP:
```bash
gcloud auth login
gcloud auth application-default login
```

First build the docker images:
```bash
./rebuild.dev.sh
```

Then run the containers:
```bash
./redeploy.dev.sh
```