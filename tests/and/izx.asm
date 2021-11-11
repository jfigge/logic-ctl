  .org $0000

  .org $0200
  lda #$ff
  ldx #$02
  and ($fa,x)
  ldx #$04
  and ($fa,x)
  ldx #$06
  and ($fa,x)
  ldx #$08
  and ($fa,x)
  ldx #$0a
  and ($fa,x)

  .org $00fc
  .word $0230
  .word $0231
  .org $0000
  .word $0232
  .word $0233
  .word $0234

  .org $0230
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
