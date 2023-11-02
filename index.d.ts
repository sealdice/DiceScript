export declare function newVM(): DiceScriptContext;
export declare function newVMForPlaygournd(): DiceScriptContext;
export declare const help: string;

export declare function vmNewDict(): VMValue;
export declare function vmNewFloat(): VMValue;
export declare function vmNewInt(): VMValue;
export declare function vmNewStr(): VMValue;
export declare function newValueMap(): ValueMap;
export declare function newConfig(): RollConfig;

declare interface GoError {
  $type: 'errors.*error';
  /** error info */
  Error(): string;
}

declare interface ValueMap {
  $type: 'github.com/sealdice/dicescript.*ValueMap';

  Load(key: string): [VMValue | null, boolean];
  MustLoad(key: string): VMValue | null;
  Store(key: string, value: VMValue): void;
  LoadOrStore(key: string, value: VMValue): [VMValue, boolean];
  LoadAndDelete(key: string): [VMValue | null, boolean];
  Delete(key: string): void;
  Range(f: (key: string, value: VMValue) => boolean): void;

  __internal_object__: {
    $val: any;
    mu: any;
    read: any;
    dirty: boolean;
    misses: number;
  };
}

declare interface VMValue {
  $type: 'github.com/sealdice/dicescript.*VMValue';

  ToJSONRaw(save: Map<VMValue, boolean>): [Uint8Array, GoError];
  ToJSON(): [Uint8Array, GoError];
  UnmarshalJSON(input: Uint8Array): GoError;
  ArrayItemGet(ctx: DiceScriptContext, index: number): VMValue | null;
  ArrayItemSet(ctx: DiceScriptContext, index: number, val: VMValue): boolean;
  ArrayFuncKeepBase(ctx: DiceScriptContext, pickNum: number, orderType: number): [boolean, number];
  ArrayFuncKeepHigh(ctx: DiceScriptContext, pickNum: number): [boolean, number];
  ArrayFuncKeepLow(ctx: DiceScriptContext, pickNum: number): [boolean, number];
  Clone(): VMValue;
  AsBool(): boolean;
  ToString(): string;
  // toStringRaw(ri: recursionInfo): string;
  // toReprRaw(ri: recursionInfo): string;
  ToRepr(): string;
  ReadInt(): [number, boolean];
  ReadFloat(): [number, boolean];
  ReadString(): [string, boolean];
  // ReadArray(): [ArrayData, boolean];
  // ReadComputed(): [ComputedData, boolean];
  // ReadDictData(): [DictData, boolean];
  // MustReadDictData(): DictData;
  // MustReadArray(): ArrayData;
  MustReadInt(): number;
  MustReadFloat(): number;
  // ReadFunctionData(): [FunctionData, boolean];
  // ReadNativeFunctionData(): [NativeFunctionData, boolean];
  // ReadNativeObjectData(): [NativeObjectData, boolean];
  OpAdd(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpSub(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpMultiply(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpDivide(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpModulus(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpPower(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpNullCoalescing(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompLT(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompLE(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompEQ(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompNE(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompGE(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpCompGT(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpBitwiseAnd(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpBitwiseOr(ctx: DiceScriptContext, v2: VMValue): VMValue;
  OpPositive(): VMValue;
  OpNegation(): VMValue;
  AttrSet(ctx: DiceScriptContext, name: string, val: VMValue): VMValue;
  AttrGet(ctx: DiceScriptContext, name: string): VMValue;
  ItemGet(ctx: DiceScriptContext, index: VMValue): VMValue;
  ItemSet(ctx: DiceScriptContext, index: VMValue, val: VMValue): boolean;
  GetSlice(ctx: DiceScriptContext, a: number, b: number, step: number): VMValue;
  Length(ctx: DiceScriptContext): number;
  GetSliceEx(ctx: DiceScriptContext, a: VMValue, b: VMValue): VMValue;
  SetSlice(ctx: DiceScriptContext, a: number, b: number, step: number, val: VMValue): boolean;
  SetSliceEx(ctx: DiceScriptContext, a: VMValue, b: VMValue, val: VMValue): boolean;
  ArrayRepeatTimesEx(ctx: DiceScriptContext, times: VMValue): VMValue;
  GetTypeName(): string;
  ComputedExecute(ctx: DiceScriptContext): VMValue;
  FuncInvoke(ctx: DiceScriptContext, params: VMValue[]): VMValue;
  FuncInvokeNative(ctx: DiceScriptContext, params: VMValue[]): VMValue;
  AsDictKey(): [string, GoError];

  __internal_object__: {
    $val: any,
    TypeId: any,
    Value: any
  };
}

declare interface RollConfig {
  EnableDiceWoD: boolean;
  EnableDiceCoC: boolean;
  EnableDiceFate: boolean;
  EnableDiceDoubleCross: boolean;
  
  DisableBitwiseOp: boolean;
  DisableStmts: boolean;
  DisableNDice: boolean;

  CallbackLoadVar: (name: string) => [string, VMValue];
  CallbackSt: (type: string, name: string, val: VMValue, extra: VMValue, op: string, detail: string) => void;

  OpCountLimit: number;
  DefaultDiceSideExpr: string;
  defaultDiceSideExprCacheFunc: VMValue;

  PrintBytecode: boolean;
  IgnoreDiv0: boolean;

  DiceMinMode: boolean;
  DiceMaxMode: boolean;
}


export declare interface DiceScriptContext {
  $type: 'github.com/sealdice/dicescript.*Context';

  RunExpr(value: string): VMValue | undefined;
  /** Eval a code, store result in ctx.Ret, store error in ctx.Error */
  Run(expr: string): void;
  GetAsmText(): string;
  StackTop(): number;
  Depth(): number;
  Init(): void;
  SetConfig(cfg: RollConfig);
  loadInnerVar(name: string): VMValue | undefined;
  LoadNameGlobal(name: string, isRaw: boolean): VMValue | undefined;
  LoadNameLocal(name: string, isRaw: boolean): VMValue | undefined;
  LoadName(name: string, isRaw: boolean): VMValue | undefined;
  StoreName(name: string, v: VMValue): void;
  StoreNameLocal(name: string, v: VMValue): void;
  StoreNameGlobal(name: string, v: VMValue): void;

  stack: VMValue[];
  top: number;
  NumOpCount: number;
  Error: GoError;
  Ret: VMValue | null;
  RestInput: string;
  Matched: string;
  Detail: string;
  IsRunning: boolean;
  readonly Config: RollConfig;
  // flagsStack: RollConfig[];
  // CustomDiceInfo: customDiceItem[];
  ValueStoreHookFunc: (ctx: DiceScriptContext, name: string, v: VMValue) => boolean;
  globalNames: ValueMap;
  GlobalValueStoreFunc: (name: string, v: VMValue) => void;
  GlobalValueLoadFunc: (name: string) => VMValue;

  __internal_object__: {
    $val: any;
    parser: any;
    subThreadDepth: number;
    attrs: any;
    upCtx: any;
  };
}
