  .org $0000

  .org $0200
  cli
  ldx #$00
  inx
  jmp $0203

  .org $0230
  ldx #$00
  rti

  .org $fffa
  .word $0230
  .word $0200
  .word $0230
