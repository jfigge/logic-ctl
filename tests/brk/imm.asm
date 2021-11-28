  .org $0000

  .org $0200
  brk
  lda #$00
  nop
  jmp $0200

  .org $0220
  rti

  .org $fffa
  .word $0220
  .word $0200
  .word $0220
