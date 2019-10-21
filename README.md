# app-store
This is the backend that provides the service behind apps.protos.io


## Dev instructions

### Build the development Docker image

```
$ docker build --target builder -t registry.gitlab.com/protosio/app-store:dev
```

### Start the docker-compose stack

```
$ APPSTORE_POSTGRES_PASSWD=<password> APPSTORE_POSTGRES_USER=<username> APP_STORE_DATA_PATH=<docker volume data path> docker-compose -f docker-compose-dev.yml up
```

### Apply the DB migrations

First enter the app-store container which has the db migration tool available:

```
$ docker exec -ti app-store_app-store_1 /bin/bash
```

Then execute the migrations:

```
root@b67d2d4436de:/go/src/github.com/protosio/app-store# migrate -source file://migrations/ -database "postgres://${APPSTORE_POSTGRES_USER}:${APPSTORE_POSTGRES_PASSWD}@postgres:5432/${APPSTORE_POSTGRES_USER}?sslmode=disable" up
1/u create_table (63.7714ms)
2/u add_tsvector (140.0916ms)
3/u add_versions_metadata (195.6604ms)
4/u update_tsvector (270.2165ms)
5/u add_installer_id (392.3968ms)
```

At this point the development app-store should be usable.

## Prod instructions
