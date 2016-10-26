# TAP Ceph Broker

Ceph broker is a microservice developed to be part of TAP platform. 
It is used to create and delete ceph RBD volumes.

## REQUIREMENTS

### Binary
* rbd (ceph utility)

### Compilation
* git (for pulling repository)
* go >= 1.6

## Compilation
To build project:
```bash
  git clone https://github.com/intel-data/tap-ceph-broker
  cd tap-ceph-broker
  make build_anywhere
```
Binary is available in ./application directory.

To build RPM:
```bash
  export PKG_VERSION=0.8
  make build_rpm
```

## USAGE

To provide IP, port and credentials for the application, you have to setup system environment variables:
```bash
export BIND_ADDRESS=127.0.0.1
export PORT=80
export CEPH_BROKER_USER=admin
export CEPH_BROKER_PASS=password
```

Ceph Broker endpoints are documented in swagger.yaml file.
Below you can find sample Ceph Broker usage.

#### Create RBD volume
To create RBD volume "test_volume" and format it with xfs filesystem type:
```bash
curl -H "Content-Type: application/json" -X POST -d '{"imageName": "test_volume", "size":1024, "fileSystem": "xfs"}' http://127.0.0.1/api/v1/rbd --user admin:password
```

#### Delete RBD volume
To delete previously created "test_volume" volume:
```bash
curl -X DELETE http://127.0.0.1/api/v1/rbd/test_volume --user admin:password
```
