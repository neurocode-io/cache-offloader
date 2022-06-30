#!/bin/bash
  
set -m
  
npm start &
./cache-offloader
  
fg %1