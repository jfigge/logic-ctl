  .org $0000
  byte $40
  .org $0200
start:
  clv
  bvs start
  jmp middle

  .org $02b0
middle:
  bit $00
  bvs edge
  jmp $0300

  .org $02fd
edge:
  bvs forward
  jmp $0300

forward:
  bvs middle

