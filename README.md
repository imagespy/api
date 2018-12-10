# ImageSpy API

## Usage

### Requirements

* A mysql database
* A Docker Registry

### Server

The Server exposes the API via HTTP in JSON format. It is a long-running process.

1. Configure Docker Registry to send events to the Server

        notifications:
          endpoints:
            - name: imagespy
              url: https://imagespy.example.com/registry/event
              timeout: 500ms
              threshold: 5
              backoff: 1s
              # Recommended: Avoids sending an event on every push of a layer
              ignoredmediatypes:
                - application/octet-stream

    Check [Working with notifications](https://docs.docker.com/registry/notifications/) in the official documentation of the official Docker Registry documentation for more information on events.

2. Start the Server:

        ./api server --db.connection "root:root@tcp(127.0.0.1:3306)/imagespy?charset=utf8&parseTime=True&loc=Local" --http.address "127.0.0.1:3001" --registry.address "registry.example.com" --registry.password "secret" --registry.username "reguser"

3. Push a new Docker image to the Registry

### Updater

The Updater is checks if a newer version of a Docker image is available at the Docker Registry and updates it. It is a one-off process and should be scheduled to run by external tools, e.g. cron.

Start the Updater:

```
./api updater --db.connection "root:root@tcp(127.0.0.1:3306)/imagespy?charset=utf8&parseTime=True&loc=Local" --registry.address "registry.example.com" --registry.password "secret" --registry.username "reguser"
```

**Note:** It is not strictly necessary to run the Updater when the Server is configured to receive events from a Docker Registry. Scheduling it to run at least once a day can still be beneficial to ensure images are up-to-date in case the Server missed events due to downtime.

## Development

### Build

```
make build
```

### Test

```
make test
```
