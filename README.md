<h1 align="center"><code>roverlib</code> for <code>go</code></h1>
<div align="center">
  <a href="https://github.com/VU-ASE/roverlib-go/releases/latest">Latest release</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ase.vu.nl/docs/category/roverlib-go">Documentation</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ase.vu.nl/docs/framework/glossary/roverlib">About the roverlib</a>
  <br />
</div>
<br/>

**When building a service that runs on the Rover and should interface the ASE framework, you will most likely want to use a [roverlib](https://ase.vu.nl/docs/framework/glossary/roverlib). This is the variant for go.**

## Initialize a Go service

You can initialize a service with the Go roverlib using `roverctl` as follows:

```bash
roverctl service init go --name go-example-service --source github.com/author/example-service-service
```

Read more about using `roverctl` to initialize services [here](https://ase.vu.nl/docs/framework/Software/rover/roverctl/usage#initialize-a-service).


