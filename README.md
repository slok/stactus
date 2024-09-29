# Stactus

Your modern static status page generator.

Are you familiar with [Jekyll](https://jekyllrb.com/), [Hugo](https://gohugo.io/), [mkdocs](https://www.mkdocs.org/) or similar generators? if yes, you already know how to use stactus.

With Stactus you can create beautiful status page with simple YAMLs in matter of seconds, the generated HTMLs can be uploaded to places like [Github pages](https://pages.github.com/) or serve them with a simple Nginx.

Creating and updating and incident is as simple as this:

```yaml
version: incident/v1
id: fnj6l879tbtw
name: Disruption with some GitHub services
impact: minor
timeline:
  - ts: "2024-08-12 14:03:13"
    description: We are currently investigating this issue.
    investigating: true
  - ts: "2024-08-12 14:15:39"
    description: This incident has been resolved.
    resolved: true
```

And will render this:

![Incident example](docs/img/readme-ir-example.png)

## Features

- Simple architecture (single binary).
- Simple and easy to use APIs based on YAML.
- No need for a backend, it's just HTML what it generates.
- Easy to deploy and operate (e.g: Github pages).
- Based on static configuration (e.g: YAML Files on a git repo).
- Customizable with built-in themes.
- Extendable with custom themes that can live on the same repo as the incidents.
- Markdown support on incident details (is optional, not required).
- Prometheus metrics (yes! they are also part of the static generation).
- Able to subscribe to updates with Atom feed (and/or Prometheus metrics).
- Atlassian status page migrator.

## Why

The purpose of a static page is to be a central place where clients can check the current status of a service. This means that a static page is not a place to know the availability, latency or service metrics.

A status page is a communication point, and for communication, a simple file with text is enough, you don't need an over complicated backend that checks latency and makes pings every N minutes.

**Low level metrics are for the company, communication and updates are for the clients.**

## Getting started

Run the development server.

```bash
git clone https://github.com/slok/stactus-showcase /tmp/stactus-test
stactus serve -i /tmp/stactus-test/showcases/github/stactus.yaml
```

### Example

- [Example with Github actions deploying to status pages](https://github.com/slok/stactus-test).

## API

### General settings

### Incident

## Subscriptions

- Atom:
- Prometheus metrics:

## Themes

### Available
### Custom theme

## Migrate from Atlassian status page
