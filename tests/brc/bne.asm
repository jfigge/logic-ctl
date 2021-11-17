  .org $0000
  .word $0001
  .org $0200
start:
  lda $00
  beq start
  jmp middle

  .org $02b0
middle:
  lda $01
  beq edge
  jmp $0300

  .org $02fd
edge:
  beq forward
  jmp $0300

forward:
  beq middle

