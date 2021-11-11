  .org $0000

  .org $0200
  lda #$ff
  ldx #$0d
  and $02f0,x
  inx
  and $02f0,x
  inx
  and $02f0,x
  inx
  and $02f0,x
  inx
  and $02f0,x

  .org $02fd
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
