  .org $0000

  .org $0200
  lda #$10
  ldx #$0e
  cmp $02f0,x
  inx
  cmp $02f0,x
  inx
  cmp $02f0,x

  .org $02fe
  .byte $11
  .byte $10
  .byte $08
