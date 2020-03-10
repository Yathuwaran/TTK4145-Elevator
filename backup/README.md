# Backup Module

The backup module is used to save the states and orders of all the elevators in case of malfunctions, such as power loss of loss of communication. The intention is that the remaining elevators can use the backup to learn about the lost elevators orders and state prior to the loss.

The module contains of two functions. The fucntions are simple read and write functions used to edit a local json-file.
