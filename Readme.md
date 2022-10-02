# EMM Trial

## Introduction
This project consists of two main components:
1. The Go Kernel to perform breadth first search on dataset.
2. Python notebook to verify our findings.

## Setup
#### First part
[Install golang](https://go.dev/doc/install).
After the installation you should be able to run EMM with
```bash
go run main.go
```
Alternatively you can first compile and then run the code
```bash
go build .
./emmTrial
```

#### Second part
[Install python](https://www.python.org/downloads/)
After this you can install the dependencies using
```shell
pip3 install -r requirements.txt
```
(It is recommended to use virtualenv, but it is not necessary)

After this you should be able to start a jupyter notebook using
```bash
jupyter notebook
```
At this point you should be able to open the notebook 
named 'VerifyFindings.ipynb' and run the code.