# Minecraft Server Management Framework

This framework will be a web management portal for creating, configuring, and managing Minecraft servers.

## Quickstart

To get started, copy the `.env-template` file over to `.env`. Next, change all of the passwords in the `.env` to ones you like.

## Production

For users in production, run the following

```
docker-compose up --build -d
```

The portal will be located at http://localhost

If you would like to stop the application, just run

```
docker-compose stop
```

## Development

For those wishing to help develop, run 

```
docker-compose -f docker-compose.dev.yml up --build -d && docker logs --follow --since 10s msmf_dev
```

The portal can be accessed on http://localhost:8080

This will spin up containers that will live reload any changes made to both the front end and backend.

In addition, for help debugging the database, a pgAdmin instance will be located at http://localhost:5050. The credentials for it are the ones set in the .env file

## TODO

- [x] Default admin account where the user sets the password
- [x] Login functionality
- [x] Referral system to allow other users to join the portal
- [x] Permission checking
- [ ] Permission management system
- [ ] Server creation
- [ ] Server configuration interface
- [ ] Server management
- [ ] Account settings
- [ ] Linking player accounts with user accounts
- [ ] Server auto backup functionality
- [ ] Server auto update functionality
- [ ] Add let's encrypt support