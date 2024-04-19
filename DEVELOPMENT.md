# Development

## Architecture
- Dashboard-API — Stateless Go module, which could be referred to as Karmada API and Kubernetes API extension.
- Web — Module containing web application written in React and Go server with some web-related logic (i.e., settings). Its main task is presenting data fetched from API through Dashboard-API module.

## Code Conventions
## Development Environment

Make sure the following software is installed and added to your path:

- [Docker](https://docs.docker.com/engine/install/)
- [Go](https://golang.org/dl/) (check the required version in [`modules/go.work`](modules/go.work))
- [Node.js](https://nodejs.org/en/download) (check the required version in [`modules/web/package.json`](modules/web/package.json))
- [Yarn](https://yarnpkg.com/getting-started/install) 

## Getting Started

Clone the repository of [karmada](https://github.com/karmada-io/karmada) outside karmda-dashboard project. Follow the step of [Install the Karmada control plane](https://github.com/karmada-io/karmada?tab=readme-ov-file#install-the-karmada-control-plane)
When your screen shows 'Local Karmada is running.', it means that the control plane and the clusters already started successfully.


Cloning the repository, install web dependencies with `cd modules/web && yarn`.
After that run `yarn dev`, open your browser with 'http://localhost:5173/cluster-manage', you can see some cluster which controlled by karmada control-plane. If everything ok, start your trip of development.


## Dependency Management
TBD

## Releases
TBD
