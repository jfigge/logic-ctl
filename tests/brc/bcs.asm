  .org $0000
  .org $0200
start:
  clc
  bcs start
  jmp middle

  .org $02b0
middle:
  sec
  bcs edge
  jmp $0300

  .org $02fd
edge:
  bcs forward
  jmp $0300

forward:
  bcs middle

