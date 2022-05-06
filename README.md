Build the code:

```
go build
```

This will create a binary named `create-cronjob`.

Deploy Postgres:

```
kubectl -n mynamespace apply -k postgres
```

Run the code to create a cronjob:

```
./create-cronjob -n mynamespace
```

Wait a few minutes for the cronjob to run (it is scheduled every 5 minutes).

(In the above commands, replace `mynamespace` with the namespace in which you want to deploy these manifests.)
