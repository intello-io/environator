# Environator #

## About ##

Environator is a tool for running commands against specific environment
variables, as per the [12-factor app model](http://12factor.net/).
Environment variables are specified via simple bash syntax profile files.
Plus, you can import environment variables from outside environments like a
Heroku app, which allows you to run local commands against the same
configuration as your remotely deployed apps!

## Installation ##

```bash
go get github.com/dailymuse/environator
pushd $GOPATH/src/github.com/dailymuse/environator
make install
popd
```

## Example Usage ##

```bash
# All your profiles will be stored in the `env` directory
mkdir -p env

# Let's make an example profile
echo 'DATABASE_URL=postgres://user:pass@$127.0.0.1:5432/db' > env/example_profile.env

# Now we can run a command against this profile. For example, to echo the new
# var:
e example_profile echo $$DATABASE_URL

# Let's create a new profile, that inherits from `example_profile.env` and
# adds new environment variables:
echo '{{ source "example_profile" }}' > env/example_profile_2.env
echo 'REDIS_URL=redis://localhost:6379/0' >> env/example_profile_2.env

# Now we can use that file. Let's print the overall environment:
e example_profile_2 env

# Let's make a third profile, this time pulling from a Heroku app!
echo '{{ heroku "my_heroku_app_name" }}' > env/my_heroku_app_name.env

# At run-time, environator will import the environment variables from
# `my_heroku_app_name`; you can now print its env like so:
e my_heroku_app_name env
```

## How It Works ##

You define profiles in the `env` directory of your project repo. Each profile
is a simple bash file specifying environment variables to export.

Environator parses these profiles using golang's 
[text/template](https://golang.org/pkg/text/template/) library, so you can use
that to add special logic to the profiles.
