## Install

### Go

https://go.dev/doc/install

### Google Cloud

https://cloud.google.com/sdk/docs/install


## JS/JSX

The JS code is all in `auto-agent/static/astro-world/jsx/`. This code gets compiled to `auto-agent/static/astro-world/js/` by running:

```sh
# cd auto-agent

# First time you need to run
# npm install 

npm run build
```

### Developing JS

When developing the JS/JSX you can have it compile the files on save by running:

```sh
npm run build:watch
```

#### Format JS

```sh
npm run format
```

#### Lint JS

I added eslint which checks for common errors in the JS code.

```sh
npm run lint
```

## To start the server

```sh
# cd <GIT ROOT>
app_devserver.py app.yml
```