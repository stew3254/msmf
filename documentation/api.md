# API Documentation

The intention of this document is to describe how to make requests to the API and what to expect
from it.

### Table of Contents

* [About Permissions](#about-permissions)
* [User Based Actions](#user-based-actions)
    * [Login](#login)
    * [Change Password](#change-password)
* [User Referrals](#user-referrals)
    * [Creating a Referral Code](#creating-a-referral-code)
    * [Using a Referral Code](#using-a-referral-code)
    * [Listing Active Referral Codes](#listing-active-referral-codes)

## About Permissions

Before reading any further documentation it must be made clear that anything requiring permissions
for an authenticated user gets overridden in the following scenarios. Therefore, if documentation
below says you need a certain permission to take an action, relevant exceptions here apply or
override those below without being explicitly stated. The rules are sequential, meaning that each
rule takes precedence over the next.

1. The Admin user created by the framework itself is a permanent administrator in
   **every** scenario. They have truly complete and full power over everything. Nobody can revoke
   their permissions. **Be careful and don't let others have access to this!**

1. The owner of a server has full control of anything at that server level regardless of permissions
   assigned to them. They are allowed to take any actions they want on a server they own.

1. If the user has the user level permission `administrator`, they have full control of everything
   to the web server and all game servers on it. Any permission checks will automatically pass for
   them. In addition, this user is able to assign and remove administrator to other users as
   well. **Be careful who you trust with this permission!**

1. If the user has a server level permission `administrator`, they have full control of everything
   at the game server level. Similar to user level permission 'administrator', you can add and
   revoke said permission from every other user in the portal for this particular game server.

## User Based Actions

This section will describe actions outside the regular `/api` endpoint. Things such as login, and
changing your password will appear here

### Login

First and foremost if logging into the server. The actual API is unavailable to the user if they do
not log in.

#### Format

You must send an HTTP POST request to `/login` with the form data containing your username and
password. If the username and password pair are valid, the server will respond with a status found
and redirect you to the login page. It will also return a token in the form of a cookie to be used
for future authentication. If the login is invalid, you will get an HTTP Status Forbidden error

#### Example

Request:

```http request
POST /login
Host: localhost:8080
Content-Type: application/x-www-form-urlencoded

username: "user"
password: "password"
```

Response:

```http request
HTTP/1.1 302 Found
Content-Length: 0
Date: Wed, 28 Jul 2021 15:10:05 GMT
Location: /
Set-Cookie: token=33dffe8cb1743d9ddfe82bd2b3caeb3510c3020ddae2f6770cbfb26103be2c32; Expires=Wed, 28 Jul 2021 21:10:05 GMT; HttpOnly; Secure; SameSite=Strict
```

### Change Password

As a user who knows their password, you can change your password by hitting this endpoint.
Currently, there does not exist a way to recover a password that is lost so don't forget it!

#### Format

You must send an HTTP POST request to `/change-password` with the form data containing your current
password, and your new password. If the current password is invalid, you will get an HTTP Status
Forbidden error. If the request is valid, the server will respond with a status found and redirect
you to the index page.

#### Example

Request:

```http request
POST /change-password
Host: localhost:8080
Content-Type: application/x-www-form-urlencoded

current_password: "bad password"
new_password: "T0tally_$uper S3cur3 P@ssword!."
```

Response:

```http request
HTTP/1.1 302 Found
Content-Length: 0
Date: Wed, 28 Jul 2021 15:11:00 GMT
Location: /
```

## User Referrals

This section will describe how to invite a user to the web portal. In order to get an account on
this framework, you must create a referral code to give to a user. This will allow them to use the
referral code to create an account themselves.

### Creating a Referral Code

To create a referral code, you must first be an authenticated user and have the permission
`invite_user`.

### Using a Referral Code

### Listing Active Referral Codes
