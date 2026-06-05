
export interface FlowInput {
  name: string | undefined
  params: string[]
}
export interface Flow extends FlowInput {
  id: number
  params_replaced: string
  run_count: number
  step_count: number
}
export function NewFlow(): Flow {
  return  {
    name: undefined,
    params: [],
    id: 0,
    params_replaced: '',
    run_count: 0,
    step_count: 0
  }
}
export function GetFlowInputFromItem(item: Flow): FlowInput {
  return  {
    name: item.name,
    params: item.params
  }
}


export interface Point {
  cmd: string
  depends_on: number[]
  display_order: number
  err_msg: string
  finished_at: Date
  id: number
  started_at: Date
  status: string
  step_name: string
}


export interface Run {
  finished_at: Date
  flow_fk: number
  flow: string
  id: number
  point_count: number
  point_stati: string
  started_at: Date
  step_id: number
  step_name: string
}


export interface StepInput {
  cmd: string
  display_order: number
  flow_fk: number | undefined
  name: string
}
export interface Step extends StepInput {
  depends_on: number[]
  depends_on_names: string[]
  id: number
}
export function NewStep(flowFk?: number): Step {
  return  {
    cmd: '',
    display_order: 0,
    flow_fk: flowFk ?? undefined,
    name: '',
    depends_on: [],
    depends_on_names: [],
    id: 0,
  }
}
export function GetStepInputFromItem(item: Step): StepInput {
  return  {
    cmd: item.cmd,
    display_order: item.display_order,
    flow_fk: item.flow_fk,
    name: item.name,
  }
}


export interface StepLinkInput {
  depends_on_fk: number | undefined
  step_fk: number | undefined
}
export interface StepLink extends StepLinkInput {
  id: number
}
export function NewStepLink(stepFk?: number): StepLink {
  return  {
    depends_on_fk: undefined,
    step_fk: stepFk ?? undefined,
    id: 0,
  }
}
export function GetStepLinkInputFromItem(item: StepLink): StepLinkInput {
  return  {
    depends_on_fk: item.depends_on_fk,
    step_fk: item.step_fk,
  }
}