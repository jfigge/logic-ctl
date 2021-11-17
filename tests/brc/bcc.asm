  .org $0000
  .org $0200
start:
  sec
  bcc start
  jmp middle

  .org $02b0
middle:
  clc
  bcc edge
  jmp $0300

  .org $02fd
edge:
  bcc forward
  jmp $0300

forward:
  bcc middle

