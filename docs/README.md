---
title: Polycrate
description: A framework to package, integrate and automate complex applications and infrastructure
---

# Polycrate

Polycrate is a framework that lets you package, integrate and automate complex applications and infrastructure. With Polycrate you can bundle dependencies, tools, cofiguration and deployment logic for any kind of IT system in a single workspace and expose reusable actions that enable a streamlined DevOps workflow.

<figure markdown>
  ![Polycrate](logo.svg){: style="height:75px;"}
  <figcaption>Polycrate</figcaption>
</figure>

If you're working with modern cloud native tooling you most likely know the pain of dependency- and tool-management, the mess of git-repositories to get Infrastructure-as-Code working and the leaking portability when it comes to working in a team. Polycrate helps you to glue all the command-line tools, configuration files, secrets and deployment scripts together, package it into a version-controlled workspace and provide a seemless way to expose well-defined actions that allow for easy replication and low-risk execution of common workflows.

Polycrate revolves around a [workspace](reference.md#workspace) of [blocks](reference.md#blocks). Blocks have [actions](reference.md#actions) that can be automated using [workflows](reference.md#workflows). This way you can build complex but flexible automations and platforms that you can easily share with your team. 

You can use Polycrate to create an overarching "platform development framework" for your organization - using shared **blocks** and **workspaces** makes it easy to establish common tactics and best-practices for deployment and operations for your team(s).

It has been created to help builders to build faster. While being developed in a cloud native spirit, Polycrate is fully capable of handling "legacy" infrastructure and operations as well. Polycrate makes it easy for individuals to manage complex IT systems with confidence. With its "Infrastructure-as-Code" approach, Polycrate makes your operational tasks portable and repeatable. 

Polycrate executes blocks in a portable container that comes with a plethora of useful tooling out of the box. Construct your infrastructure with Terraform, automate your deployments with Ansible and manage your containers with Kubernetes. Polycrate works on all major operating systems and comes with only a single dependency - **Docker**.



[Get started](getting-started.md){ .md-button .md-button--primary }
[:simple-discord: Discord](https://discord.gg/8cQZfXWeXP){ .md-button }