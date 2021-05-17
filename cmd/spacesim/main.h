//+build windows darwin
#include <stdint.h>

typedef struct BodyInfo {
  short Id;
  short Owner;
  int Size;
  int Proximity;
  int Lifetime;
  float BounceCoefficient;
  float VelocityX;
  float VelocityY;
} BodyInfo;
