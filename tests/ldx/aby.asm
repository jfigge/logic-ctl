  .org $0000
  .org $0200
  ldy #$20
  ldx $0200,y
  ldx $02f0,y

  .org $0220
  .byte $10
  .org $0310
  .byte $11
