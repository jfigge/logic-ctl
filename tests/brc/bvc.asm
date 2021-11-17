  .org $0000
  byte $40
  .org $0200
start:
  bit $00
  bvc start
  jmp middle

  .org $02b0
middle:
  clv
  bvc edge
  jmp $0300

  .org $02fd
edge:
  bvc forward
  jmp $0300

forward:
  bvc middle

