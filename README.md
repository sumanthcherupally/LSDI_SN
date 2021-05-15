# LSDI (Storage Node)

LSDI (lightweight and scalable DAG based blockchain for verifying IoT data integrity).


## Table of contents
- Background
- Usage
    - Prerequisites
    - Build

## Background

This repository has a implementation of LSDI storage node. LSDI is a DAG based blockchain system inspired from IoTA to improve the scalabilty of using blockchain in IoT environments. To find more about the system please read https://ieeexplore.ieee.org/abstract/document/9334000/.

## Usage

### Prerequisites

- golang (version > 1.08) installed
- Discovery node up and running
- Here is the [gateway node repository](https://github.com/srinivasboga7/LSDI_GW)

### Build

You can build the node by running :
> go build storage.go

- Detailed instructions to run the complete system is [here](https://docs.google.com/document/d/1d1XgUyclQzoNh-tGMhnTrxbjGm4oyc-V916PS11_baw/edit?usp=sharing)

