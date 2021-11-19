  .org $0000
  byte $02

  .org $0200
  jmp ($0300)

  .org $0220
  jmp ($0302)

  .org $0240
  jmp ($0304)

  .org $0260
  jmp ($00ff)


  .org $0300
  .word $0220
  .word $0240
  .word $0260

  .org $00ff
  .word $0300
