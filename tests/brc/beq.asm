  .org $0000
  .word $0100
  .org $0200
start:
  lda $00
  bne start
  jmp middle

  .org $02b0
middle:
  lda $01
  bne edge
  jmp $0300

  .org $02fd
edge:
  bne forward
  jmp $0300

forward:
  bne middle

