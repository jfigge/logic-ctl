  .org $0000

  .org $0200
  lda #$ff
  ldy #$0d
  and $02f0,y
  iny
  and $02f0,y
  iny
  and $02f0,y
  iny
  and $02f0,y
  iny
  and $02f0,y

  .org $02fd
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
