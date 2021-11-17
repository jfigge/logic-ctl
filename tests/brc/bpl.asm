  .org $0000
  .org $0200
start:
  ldx #$80
  bpl start
  jmp middle

  .org $02b0
middle:
  ldx #$00
  bpl edge
  jmp $0300

  .org $02fd
edge:
  bpl forward
  jmp $0300

forward:
  bpl middle

