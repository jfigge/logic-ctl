  .org $0000
  .org $0200
  ldx #$20
  ldy $0200,x
  ldy $02f0,x

  .org $0220
  .byte $10
  .org $0310
  .byte $11
