# Bill.com Mock Service

A simple API mock server that that returns examples specified in an API description document. Features include:

- Authentication
    - Login
    - Logout
- Search List
    - ActgClass
    - Vendor
- Vendor Management
    - Vendor
        - Create
        - Read
        - Update
- Transactions      
    - Bill
        - Read
        - Create
        - Delete
        - Update
## Usage is simple
```bash
mockserver --port 8080
```

## Docker Image

Docker makes it easy to run locally. For example:

```bash
docker build -t mock.bill.com:latest
docker run -i -t -p 8080:8080 mock.bill.com:latest:latest
```
