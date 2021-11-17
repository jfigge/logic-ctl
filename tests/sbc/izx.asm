  .org $0000

  .org $0200
  sec
  lda #$ff
  ldx #$02
  sbc ($fa,x)
  ldx #$04
  sbc ($fa,x)
  ldx #$06
  sbc ($fa,x)
  ldx #$08
  sbc ($fa,x)
  ldx #$0a
  sbc ($fa,x)

  .org $00fc
  .word $0230
  .word $0231
  .org $0000
  .word $0232
  .word $0233
  .word $0234

  .org $0230
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
